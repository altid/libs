package server

import (
	"context"
	"log"

	"github.com/altid/server/client"
	"github.com/altid/server/files"
	"github.com/altid/server/internal/services"
)

type Runner interface {
	// Run will be called internally once services have been iterated and set up
	// name will be the name of the service, and initial will be the default buffer
	// Client should be initialized `c.Client(0)` and set to `initial`
	Run(context.Context, *Service) error
	Address() (string, string)
}

type Server struct {
	// Logger can be set to log server messages
	// such as buffer activity or control message writes
	Logger   func(string, ...interface{})
	ctx      context.Context
	services map[string]*services.Service
	run      Runner
}

type Service struct {
	Client *client.Manager
	Files  *files.Files
	Name   string
	Buffer string // Default buffer
}

// NewServer creates a server to manage services
func NewServer(ctx context.Context, runner Runner, dir string) (*Server, error) {
	services, err := services.FindServices(ctx, dir)
	if err != nil {
		return nil, err
	}

	s := &Server{
		ctx:      context.Background(),
		run:      runner,
		services: services,
	}

	return s, nil
}

// Listen wraps calls to the Service's Run method
// it will return on unrecoverable errors
func (s *Server) Listen() error {
	errs := make(chan error)

	for _, svc := range s.services {
		go func(svc *services.Service) {
			//addr, port := s.run.Address()
			//s.Logger("using port %d", port)
			//m := &mdns.Entry{
			//	Addr: addr,
			//	Name: svc.Name,
			//	Port: port,
			//}

			//if e := mdns.Register(m); e != nil {
			//	s.Logger("%v", e)
			//}

			service := &Service{
				Name:   svc.Name,
				Client: svc.Client,
				Files:  svc.Files,
				Buffer: svc.Tabs.Default(),
			}

			err := s.run.Run(s.ctx, service)
			// TODO: switch on recoverable errors
			// as we build out the library
			switch err {
			default:
				log.Printf("%v", err)
				errs <- err
			}
		}(svc)
	}

	select {
	case err := <-errs:
		return err
	case <-s.ctx.Done():
		return nil
	}
}
