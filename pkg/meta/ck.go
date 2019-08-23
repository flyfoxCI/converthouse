package meta

import (
	"time"

	"github.com/fagongzi/util/json"
	"github.com/infinivision/prophet"
)

// CKInstance is the clickhourse metadata
type CKInstance struct {
	InstanceID uint64         `json:"id"`
	Addr       string         `json:"addr"`    // Shard addr
	ClientAddr string         `json:"cliAddr"` // CK addr
	Action     prophet.Action `json:"action"`
	NodeLabels []prophet.Pair `json:"-"`
}

// ShardAddr shard addr
func (ck *CKInstance) ShardAddr() string {
	return ck.Addr
}

// SetID adapter method
func (ck *CKInstance) SetID(id uint64) {
	ck.InstanceID = id
}

// ID adapter method
func (ck *CKInstance) ID() uint64 {
	return ck.InstanceID
}

// Labels adapter method
func (ck *CKInstance) Labels() []prophet.Pair {
	return ck.NodeLabels
}

// State adapter method
func (ck *CKInstance) State() prophet.State {
	return prophet.UP
}

// ActionOnJoinCluster adapter method
func (ck *CKInstance) ActionOnJoinCluster() prophet.Action {
	return ck.Action
}

// Clone adapter method
func (ck *CKInstance) Clone() prophet.Container {
	value := &CKInstance{}
	json.MustUnmarshal(value, json.MustMarshal(ck))
	value.NodeLabels = ck.NodeLabels
	return value
}

// Marshal adapter method
func (ck *CKInstance) Marshal() ([]byte, error) {
	return json.MustMarshal(ck), nil
}

// Unmarshal adapter method
func (ck *CKInstance) Unmarshal(data []byte) error {
	json.MustUnmarshal(ck, data)
	return nil
}

// TableShard is a CK distributed table sharding, it's a resource at prophet.
// Every TableShard has some replicas on different CKInstances.
// And the prophet will rebalance the TableShard replicas.
type TableShard struct {
	ResID        uint64         `json:"id"`
	Table        string         `json:"table"`
	ShardPeers   []prophet.Peer `json:"peers"`
	ShardLabels  []prophet.Pair `json:"labels"`
	Version      uint64         `json:"version"`
	ScaleVersion uint64         `json:"scaleVersion"`
	Partitions   []uint64       `json:"partitions"`
	CreatedBy    uint64         `json:"createdBy"`
	Offset       uint64         `json:"offset"`

	CheckTime time.Time `json:"-"`
	HasCheck  bool      `json:"-"`
}

// SetID adapter method
func (t *TableShard) SetID(id uint64) {
	t.ResID = id
}

// ID adapter method
func (t *TableShard) ID() uint64 {
	return t.ResID
}

// Peers adapter method
func (t *TableShard) Peers() []*prophet.Peer {
	var peers []*prophet.Peer
	for _, peer := range t.ShardPeers {
		p := peer
		peers = append(peers, &p)
	}
	return peers
}

// Labels adapter method
func (t *TableShard) Labels() []prophet.Pair {
	return t.ShardLabels
}

// SetPeers adapter method
func (t *TableShard) SetPeers(peers []*prophet.Peer) {
	var values []prophet.Peer
	for _, peer := range peers {
		values = append(values, *peer)
	}
	t.ShardPeers = values
}

// ScaleCompleted adapter method
func (t *TableShard) ScaleCompleted(id uint64) bool {
	return t.CreatedBy >= id || len(t.Partitions) < 2
}

// Stale adapter method
func (t *TableShard) Stale(other prophet.Resource) bool {
	otherT := other.(*TableShard)
	return otherT.Version < t.Version
}

// Changed adapter method
func (t *TableShard) Changed(other prophet.Resource) bool {
	otherT := other.(*TableShard)
	return otherT.Version > t.Version ||
		otherT.ScaleVersion > t.Version ||
		otherT.CreatedBy > t.CreatedBy
}

// Clone adapter method
func (t *TableShard) Clone() prophet.Resource {
	value := &TableShard{}
	json.MustUnmarshal(value, json.MustMarshal(t))
	return value
}

// Marshal adapter method
func (t *TableShard) Marshal() ([]byte, error) {
	return json.MustMarshal(t), nil
}

// Unmarshal adapter method
func (t *TableShard) Unmarshal(data []byte) error {
	json.MustUnmarshal(t, data)
	return nil
}

// HasGap has gap partition
func (t *TableShard) HasGap(other *TableShard) bool {
	for _, part := range t.Partitions {
		for _, otherPart := range other.Partitions {
			if part == otherPart {
				return true
			}
		}
	}

	return false
}
