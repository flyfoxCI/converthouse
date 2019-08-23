package server

import (
	"github.com/deepfabric/converthouse/pkg/store"
)

// Cfg server cfg
type Cfg struct {
	AddrHTTP string    `json:"addrHTTP"`
	Store    store.Cfg `json:"store"`
}
