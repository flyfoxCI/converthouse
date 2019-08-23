package server

import (
	"github.com/deepfabric/converthouse/pkg/ck"
	"github.com/deepfabric/converthouse/pkg/store"
)

// Server server
type Server struct {
	cfg Cfg

	store store.Store
}

// NewServer create a server
func NewServer(cfg Cfg) *Server {
	s := new(Server)
	s.cfg = cfg

	s.store = store.NewStore(cfg.Store, ck.NewMemCKAPI())
	return s
}

// Start start the server, include http server and prophet
func (s *Server) Start() {
	s.store.Start()
}

// Stop stop
func (s *Server) Stop() {
	s.store.Stop()
}
