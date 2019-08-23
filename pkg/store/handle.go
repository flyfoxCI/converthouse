package store

import (
	"fmt"
	"time"

	"github.com/fagongzi/goetty"

	"github.com/deepfabric/converthouse/pkg/meta"
	"github.com/fagongzi/log"
	"github.com/infinivision/prophet"
)

func (s *store) CreateTableShard(shard *meta.TableShard) error {
	var gapShard *meta.TableShard
	s.resStore.ForeachReplica(func(pr *prophet.PeerReplica) bool {
		ts := pr.Resource().(*meta.TableShard)
		if ts.Table == shard.Table && ts.HasGap(shard) {
			gapShard = ts
			return false
		}

		return true
	})

	if gapShard != nil {
		return fmt.Errorf("has gap of %+v", gapShard.ResID)
	}

	shard.ResID = s.localStore.MustAllocID()
	shard.CreatedBy = shard.ResID
	shard.ShardPeers = append(shard.ShardPeers,
		prophet.Peer{
			ID:          s.localStore.MustAllocID(),
			ContainerID: s.meta.InstanceID,
		})
	newPR := prophet.NewPeerReplica(s.resStore, shard, shard.ShardPeers[0], s, s.elector)
	s.resStore.AddReplica(newPR)
	return nil
}

func (s *store) AddPeer(res prophet.Resource, peer prophet.Peer) {
	shard := res.(*meta.TableShard)
	shard.ShardPeers = append(shard.ShardPeers, peer)
	shard.Version++

	maxID := peer.ID
	if maxID < peer.ContainerID {
		maxID = peer.ContainerID
	}

	if maxID > shard.CreatedBy {
		shard.CreatedBy = maxID
	}
}

func (s *store) RemovePeer(res prophet.Resource, peer prophet.Peer) bool {
	shard := res.(*meta.TableShard)
	removed := false
	var values []prophet.Peer
	for _, p := range shard.ShardPeers {
		if p.ID != peer.ID {
			values = append(values, p)
		} else {
			removed = true
		}
	}

	if removed {
		shard.ShardPeers = values
		shard.Version++
	}

	return removed
}

func (s *store) Scale(res prophet.Resource, data interface{}) (bool, []*prophet.PeerReplica) {
	shard := res.(*meta.TableShard)
	containerID := data.(uint64)

	// Check the scale operation have completed
	if shard.ScaleCompleted(containerID) {
		shard.CreatedBy = containerID
		return true, nil
	}

	// Only one partition, no scale required
	if len(shard.Partitions) < 2 {
		return false, nil
	}

	newShardID := s.localStore.MustAllocID()
	key := scaleKey(shard.ResID, containerID)

	succeed, exists, err := s.pd.GetStore().PutIfNotExists(key, goetty.Uint64ToBytes(newShardID))
	if err != nil {
		log.Fatalf("%+v create scale opt failed with %+v", res, err)
	}

	// At this time, the previous leader may have completed the scale operation.
	// It needs to check whether the new shard has been successfully created.
	// If it has been created, only modify the partition range of the current shard,
	// otherwise perform the Scale operation again.

	createdNewPR := succeed
	if !succeed {
		// It needs to check after 10 times of heartbeat whether the new shard has been successfully created.
		// Because the new shard will puts the metadata on the prophet at the first heartbeat.
		if !shard.HasCheck {
			shard.CheckTime = time.Now().Add(time.Second * 10)
			shard.HasCheck = true
		}

		if !time.Now().After(shard.CheckTime) {
			return false, nil
		}

		newShardID = goetty.Byte2UInt64(exists)
		target, err := s.pd.GetStore().GetResource(newShardID)
		if err != nil {
			log.Fatalf("%+v check scaled resource failed with %+v", target, err)
		}

		if target == nil {
			createdNewPR = true
		}
	}

	partitions := shard.Partitions
	point := len(shard.Partitions) / 2

	// Modify current shard
	shard.CreatedBy = containerID
	shard.Partitions = partitions[0:point]
	shard.Version++
	shard.ScaleVersion++

	if !createdNewPR {
		return true, nil
	}

	index := 0
	newShard := shard.Clone().(*meta.TableShard)
	newShard.ShardPeers = nil
	newShard.Version += 1000 // Ensure that the new shard created by the previous leader will stale
	newShard.ResID = newShardID
	for idx, p := range shard.ShardPeers {
		if p.ContainerID == s.meta.InstanceID {
			index = idx
		}

		// Assign peers on the same store to avoid large numbers of shard moving data between stores
		newShard.ShardPeers = append(newShard.ShardPeers, prophet.Peer{
			ID:          s.localStore.MustAllocID(),
			ContainerID: p.ContainerID,
		})
	}
	newShard.Partitions = partitions[point:]

	return true, []*prophet.PeerReplica{
		prophet.NewPeerReplica(s.resStore, newShard, newShard.ShardPeers[index], s, s.elector),
	}
}

func (s *store) Heartbeat(from prophet.Resource) bool {
	pr := s.resStore.GetPeerReplica(from.ID(), false)
	if pr == nil {
		log.Fatalf("bug: pr must be exists")
	}

	shard := pr.Resource().(*meta.TableShard)
	fromShard := from.(*meta.TableShard)

	peerChanged := fromShard.Version > shard.Version
	partitionChanged := fromShard.ScaleVersion > shard.ScaleVersion

	// peer changed
	if partitionChanged || peerChanged {
		shard.Version = fromShard.Version
		shard.ShardPeers = fromShard.ShardPeers
		shard.ShardLabels = fromShard.ShardLabels
		shard.CreatedBy = fromShard.CreatedBy
	}

	if partitionChanged {
		var removedPartition []uint64
		for _, p := range shard.Partitions {
			found := false
			for _, fp := range fromShard.Partitions {
				if fp == p {
					found = true
				}
			}

			if !found {
				removedPartition = append(removedPartition, p)
			}
		}

		shard.ScaleVersion = fromShard.ScaleVersion
		shard.Partitions = fromShard.Partitions
		s.mustSaveRemovedPartitions(shard.Table, removedPartition...)
	}

	return partitionChanged || peerChanged
}

func (s *store) Destory(res prophet.Resource) {
	shard := res.(*meta.TableShard)
	err := s.api.Remove(shard.Table, shard.Partitions...)
	if err != nil {
		log.Fatalf("remove table %s partition %+v from CK failed with %+v",
			shard.Table,
			shard.Partitions,
			err)
	}
}

func (s *store) ResourceBecomeLeader(res prophet.Resource) {

}

func (s *store) ResourceBecomeFollower(res prophet.Resource) {

}

func scaleKey(id, containerID uint64) string {
	return fmt.Sprintf("/converthouse/opts/scale/%d-%d", id, containerID)
}
