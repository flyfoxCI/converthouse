package store

import (
	"fmt"
	"time"

	"github.com/infinivision/prophet"
)

// Cfg server cfg
type Cfg struct {
	Name     string `json:"name"`
	DataPath string `json:"DataPath"`

	AddrReplica string         `json:"addrReplica"`
	AddrCK      string         `json:"addrCK"`
	Labels      []prophet.Pair `json:"labels"`

	MaxPeerDownTime            int `json:"maxPeerDownDuration"`
	ResourceHeartbeatInterval  int `json:"resourceHeartbeatInterval"`
	ContainerHeartbeatInterval int `json:"containerHeartbeatInterval"`

	ResourceWorkerCount uint64 `json:"resourceWorkerCount"`
}

// GetMaxPeerDownTime returns the duration value
func (c *Cfg) GetMaxPeerDownTime() time.Duration {
	return time.Second * time.Duration(c.MaxPeerDownTime)
}

// GetResourceHeartbeatInterval returns the duration value
func (c *Cfg) GetResourceHeartbeatInterval() time.Duration {
	return time.Second * time.Duration(c.ResourceHeartbeatInterval)
}

// GetContainerHeartbeatInterval returns the duration value
func (c *Cfg) GetContainerHeartbeatInterval() time.Duration {
	return time.Second * time.Duration(c.ContainerHeartbeatInterval)
}

// ProphetDir returns prophet data dir
func (c *Cfg) ProphetDir() string {
	return fmt.Sprintf("%s/prophet", c.DataPath)
}

// LocalReplicaDir returns local replica meta data dir
func (c *Cfg) LocalReplicaDir() string {
	return fmt.Sprintf("%s/replica", c.DataPath)
}
