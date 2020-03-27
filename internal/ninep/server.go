package ninep

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/altid/server/mdns"
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

	for _, svc := range s.services {
		wg.Add(1)

		go func(svc *service) {
			defer wg.Done()

			port, err := strconv.Atoi(svc.port)
			if err != nil {
				s.debug("%v (using default 564)", err)
				svc.port = "564"
				port = 564
			}

			m := &mdns.Entry{
				Addr: svc.listen,
				Name: svc.name,
				Port: port,
			}

			if e := mdns.Register(m); e != nil {
				s.debug("%v", e)
			}

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
	l := log.New(os.Stderr, "9pd server ", 0)
	l.Printf(format, v...)
}
