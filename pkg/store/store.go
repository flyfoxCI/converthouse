package store

import (
	"sync"

	"github.com/deepfabric/converthouse/pkg/ck"
	"github.com/deepfabric/converthouse/pkg/meta"
	"github.com/deepfabric/converthouse/pkg/storage"
	"github.com/fagongzi/log"
	"github.com/infinivision/prophet"
)

// Store the replica store
type Store interface {
	Start()
	Stop()

	// CreateTableShard create a new shard for Table.
	// returns error if the partitions has the gap with the exists Table shards.
	CreateTableShard(*meta.TableShard) error
}

type store struct {
	sync.RWMutex

	cfg        Cfg
	meta       meta.CKInstance
	pd         prophet.Prophet
	bootOnce   *sync.Once
	pdStartedC chan struct{}

	adapter    prophet.Adapter
	localDB    prophet.LocalStorage
	localStore prophet.LocalStore
	resStore   prophet.ResourceStore
	elector    prophet.Elector

	api ck.API
}

// NewStore creates a replica store
func NewStore(cfg Cfg, api ck.API) Store {
	localDB, err := storage.NewBadgerStorage(cfg.LocalReplicaDir())
	if err != nil {
		log.Fatalf("create local storage at %s failed with %+v", cfg.LocalReplicaDir(), err)
	}

	s := new(store)
	s.cfg = cfg
	s.meta.Addr = cfg.AddrReplica
	s.meta.ClientAddr = cfg.AddrCK
	s.meta.NodeLabels = cfg.Labels
	s.meta.Action = prophet.ScaleOutAction
	s.api = api

	s.bootOnce = &sync.Once{}
	s.localDB = localDB

	return s
}

func (s *store) Start() {
	s.startProphet()
	log.Infof("begin to start store %d", s.meta.InstanceID)

	s.startShards()
	log.Infof("all shards started")
}

func (s *store) Stop() {

}

func (s *store) startProphet() {
	log.Infof("begin to start prophet")

	s.adapter = newProphetAdapter(s)
	s.pdStartedC = make(chan struct{})
	options := append(prophet.ParseProphetOptionsWithPath(s.cfg.Name, s.cfg.ProphetDir()),
		prophet.WithScaleOnNewStore(),
		prophet.WithRoleChangeHandler(s))
	s.pd = prophet.NewProphet(s.cfg.Name, s.adapter, options...)
	s.pd.Start()
	<-s.pdStartedC
}

func (s *store) startShards() {
	s.localStore.MustLoadResources(func(value []byte) (uint64, error) {
		shard := &meta.TableShard{}
		err := shard.Unmarshal(value)
		if err != nil {
			return 0, err
		}

		s.resStore.AddReplica(prophet.NewPeerReplica(s.resStore, shard, shard.ShardPeers[0], s, s.elector))
		return shard.ResID, nil
	})
}

func (s *store) mustSaveRemovedPartitions(table string, parititions ...uint64) {
	// TODO: impl
}
