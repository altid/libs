package server

import (
	"context"
	"log"
	"strconv"

	"github.com/altid/server/files"
	"github.com/altid/server/internal/mdns"
	"github.com/altid/server/internal/services"
	"github.com/altid/server/settings"
)

type Runner interface {
	Run(context.Context, *files.Files) error
}

type Server struct {
	// Logger can be set to log server messages
	// such as buffer activity or control message writes
	Logger   func(string, ...interface{})
	ctx      context.Context
	services map[string]*services.Service
	run      Runner
}

// NewServer creates a server to manage services
func NewServer(ctx context.Context, runner Runner) *Server {
	return &Server{
		ctx: context.Background(),
		run: runner,
	}
}

// Config prepares a server for listening
// It will call the Service Config method
func (s *Server) Config(settings *settings.Settings) error {
	svcs, err := services.FromSettings(s.ctx, settings)
	if err != nil {
		return err
	}

	s.services = svcs
	return nil
}

// Listen wraps calls to the Service's Run method
// it will return on unrecoverable errors
func (s *Server) Listen() error {
	errs := make(chan error)

	for _, svc := range s.services {
		go func(svc *services.Service) {
			port, err := strconv.Atoi(svc.Port)
			if err != nil {
				svc.Port = "9000"
				port = 9000
			}

			s.Logger("using port %d", port)
			m := &mdns.Entry{
				Addr: svc.Listen,
				Name: svc.Name,
				Port: port,
			}

			if e := mdns.Register(m); e != nil {
				s.Logger("%v", e)
			}

			err = svc.Run(s.run)
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
