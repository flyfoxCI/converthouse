package store

import (
	"time"

	"github.com/deepfabric/converthouse/pkg/meta"
	"github.com/deepfabric/converthouse/pkg/util"
	"github.com/fagongzi/log"
	"github.com/infinivision/prophet"
)

// ProphetBecomeLeader this node is become prophet leader
func (s *store) ProphetBecomeLeader() {
	log.Infof("*********BecomeLeader prophet*********")
	s.bootOnce.Do(func() {
		s.doBootstrapCluster()
		s.pdStartedC <- struct{}{}
	})
	log.Infof("*********BecomeLeader prophet complete*********")
}

// BecomeFollower this node is become prophet follower
func (s *store) ProphetBecomeFollower() {
	log.Infof("*********BecomeFollower prophet*********")
	s.bootOnce.Do(func() {
		s.doBootstrapCluster()
		s.pdStartedC <- struct{}{}
	})
	log.Infof("*********BecomeFollower prophet complete*********")
}

func (s *store) doBootstrapCluster() {
	s.localStore = prophet.NewLocalStore(&s.meta, s.localDB, s.pd)
	elector, err := prophet.NewElector(s.pd.GetEtcdClient(),
		prophet.WithLeaderLeaseSeconds(5),
		prophet.WithLockIfBecomeLeader(true))
	if err != nil {
		log.Fatalf("create shard elector failed with %+v", err)
	}
	s.elector = elector
	s.resStore = prophet.NewResourceStore(&s.meta,
		s.localStore,
		s.pd,
		s.elector,
		s,
		s.adapter.NewResource,
		s.cfg.GetResourceHeartbeatInterval(),
		s.cfg.ResourceWorkerCount)
	s.resStore.Start()
}

type prophetAdapter struct {
	s *store
}

func newProphetAdapter(s *store) prophet.Adapter {
	return &prophetAdapter{
		s: s,
	}
}

func (pa *prophetAdapter) NewResource() prophet.Resource {
	return &meta.TableShard{}
}

func (pa *prophetAdapter) NewContainer() prophet.Container {
	return &meta.CKInstance{}
}

func (pa *prophetAdapter) FetchLeaderResources() []uint64 {
	var values []uint64
	pa.s.resStore.ForeachReplica(func(pr *prophet.PeerReplica) bool {
		if pr.IsLeader() {
			values = append(values, pr.Resource().ID())
		}

		return true
	})

	return values
}

func (pa *prophetAdapter) FetchResourceHB(id uint64) *prophet.ResourceHeartbeatReq {
	pr := pa.s.resStore.GetPeerReplica(id, true)
	if pr == nil {
		return nil
	}

	var err error
	req := new(prophet.ResourceHeartbeatReq)
	pr.Do(func(e error) {
		if e != nil {
			err = e
			return
		}
		req.Resource = pr.Resource().Clone()
		req.LeaderPeer = pr.Peer()
		req.PendingPeers = pr.CollectPendingPeers()
		req.DownPeers = pr.CollectDownPeers(pa.s.cfg.GetMaxPeerDownTime())
		req.ContainerID = pa.s.meta.InstanceID
	}, time.Minute)

	if err != nil {
		log.Errorf("%s fetch resource heartbeat failed with %+v", pr.Tag(), err)
		return nil
	}

	return req
}

func (pa *prophetAdapter) FetchContainerHB() *prophet.ContainerHeartbeatReq {
	// maybe bootstrap not complete
	if pa.s.resStore == nil {
		return nil
	}

	replicaCnt := uint64(0)
	leaderCnt := uint64(0)

	pa.s.resStore.ForeachReplica(func(pr *prophet.PeerReplica) bool {
		if pr.IsLeader() {
			leaderCnt++
		}

		replicaCnt++
		return true
	})

	stats, err := util.DiskStats(pa.s.cfg.DataPath)
	if err != nil {
		log.Errorf("fetch store heartbeat at %s failed with %+v",
			pa.s.cfg.DataPath,
			err)
		return nil
	}

	req := new(prophet.ContainerHeartbeatReq)
	req.Container = pa.s.meta.Clone()
	req.StorageCapacity = stats.Total
	req.StorageAvailable = stats.Free
	req.LeaderCount = replicaCnt
	req.ReplicaCount = replicaCnt

	return req
}

func (pa *prophetAdapter) ResourceHBInterval() time.Duration {
	return pa.s.cfg.GetResourceHeartbeatInterval()
}

func (pa *prophetAdapter) ContainerHBInterval() time.Duration {
	return pa.s.cfg.GetResourceHeartbeatInterval()
}

func (pa *prophetAdapter) HBHandler() prophet.HeartbeatHandler {
	return pa
}

func (pa *prophetAdapter) ChangeLeader(resourceID uint64, newLeader *prophet.Peer) {
	pr := pa.s.resStore.GetPeerReplica(resourceID, true)
	if pr != nil {
		log.Infof("%s schedule change leader to peer %+v ",
			pr.Tag(),
			newLeader)
		pr.ChangeLeaderTo(newLeader.ID, nil)
	}
}

func (pa *prophetAdapter) ChangePeer(resourceID uint64, peer *prophet.Peer, changeType prophet.ChangePeerType) {
	pr := pa.s.resStore.GetPeerReplica(resourceID, true)
	if pr != nil {
		if changeType == prophet.AddPeer {
			log.Infof("%s schedule add peer %+v", pr.Tag(), peer)
			pr.AddPeer(*peer)
		} else if changeType == prophet.RemovePeer {
			log.Infof("%s schedule remove peer %+v", pr.Tag(), peer)
			pr.RemovePeer(*peer)
		}
	}
}

func (pa *prophetAdapter) ScaleResource(resourceID uint64, byContainerID uint64) {
	pr := pa.s.resStore.GetPeerReplica(resourceID, true)
	if pr != nil {
		log.Infof("%s schedule scale by container %d",
			pr.Tag(),
			byContainerID)
		pr.Scale(byContainerID)
	}
}
