package tommyqueue

import (
	"sync"

	"github.com/jemmyh/tommyqueue/internal/broker"
	"github.com/jemmyh/tommyqueue/internal/structs"
)

type Server struct {
	logger Logger
	broker broker.CtxBroker
	status *structs.ServerStatus

	wg *sync.WaitGroup

	// all workers
	processor Worker // must have
}

func NewServer(ops ...ServerOption) *Server {
	s := new(Server)
	for _, fn := range ops {
		fn(s)
	}

	// TODO
	return s
}

type ServerOption func(s *Server)

func WithLogger(l Logger) ServerOption {
	return func(s *Server) {
		s.logger = l
	}
}

func WithBroker(b broker.CtxBroker) ServerOption {
	return func(s *Server) {
		s.broker = b
	}
}
