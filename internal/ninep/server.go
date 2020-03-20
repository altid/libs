package ninep

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
)

type Server struct {
	ctx      context.Context
	services map[string]*service
	debug    func(string, ...interface{})
	sync.Mutex
}

func NewServer(ctx context.Context, s *Settings) (*Server, error) {
	server := &Server{
		ctx:      ctx,
		services: s.services,
		debug:    serverDebugLog,
	}

	if len(server.services) < 1 {
		return nil, errors.New("no services found, exiting")
	}

	if !s.debug {
		server.debug = func(string, ...interface{}) {}
	}

	return server, nil
}

// Run until an error or all services exit
func (s *Server) Run() error {
	var wg sync.WaitGroup
	var err error

	for _, svc := range s.services {
		wg.Add(1)

		go func(svc *service) {
			defer wg.Done()
			if err = svc.run(); err != nil {
				s.debug("%v", err)
				return
			}
		}(svc)
	}

	wg.Wait()
	s.ctx.Done()

	return nil
}

func serverDebugLog(format string, v ...interface{}) {
	l := log.New(os.Stderr, "", 0)
	l.Printf(format, v...)
}
