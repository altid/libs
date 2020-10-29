package server

import (
	"context"
	"log"
	"fmt"
	"strconv"

	"github.com/altid/server/client"
	"github.com/altid/server/command"
	"github.com/altid/server/files"
	"github.com/altid/server/internal/services"
	"github.com/altid/server/internal/mdns"
)

type Runner interface {
	// Run will be called internally once services have been iterated and set up
	// name will be the name of the service, and initial will be the default buffer
	// Client should be initialized `c.Client(0)` and set to `initial`
	Run(context.Context, *Service) error
	// Address must return the IP + Port pair the server uses, for use with mDNS broadcasts internally
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
	// The client manager is expected to be used whenever a client connects to a server
	// generally a server only has to call "c := client.Client(0)"
	// This adds the client to the stack using the special ID of 0
	// making sure to client.Remove(c.UUID) after you're done with it.
	// Clients will only receive events when they are registered!
	Client *client.Manager
	// Files holds the handlers to our internal file representations
	// The Normal and Stat methods return useful types
	// Generally, Normal returns an io.ReadWriter; but you should switch on all variants
	// Of io.Writer, io.Reader, io.ReaderAt, io.WriterAt, io.Seeker, and any combination therein
	// Additionally, if the file requested is a directory, a []*os.FileInfo will be returned instead
	// Stat will return an *os.FileInfo for a given file
	Files    *files.Files
	Commands chan *command.Command
	name     string
	buffer   string
}

// NewServer returns a server which watches the event files at `dir`
// the Runner will be called for each Service that is located, facilitating
// server access to underlying state
func NewServer(ctx context.Context, runner Runner, dir string) (*Server, error) {
	services, err := services.FindServices(ctx, dir)
	if err != nil {
		return nil, err
	}

	s := &Server{
		ctx:      context.Background(),
		run:      runner,
		services: services,
		Logger:   func(string, ...interface{}) {},
	}

	return s, nil
}

// Listen wraps calls to the Service's Run method
// it will return on unrecoverable errors
func (s *Server) Listen() error {
	errs := make(chan error)

	for _, svc := range s.services {
		svc.Debug = s.Logger

		go func(svc *services.Service) {
			addr, port := s.run.Address()
			s.Logger("using port %d", port)
			pint, _ := strconv.Atoi(port)
			m := &mdns.Entry{
				Addr: addr,
				Name: svc.Name,
				Port: pint,
			}

			if e := mdns.Register(m); e != nil {
				s.Logger("%v", e)
			}

			service := &Service{
				Commands: svc.Command,
				Client:   svc.Client,
				Files:    svc.Files,
				name:     svc.Name,
				buffer:   svc.Tabs.Default(),
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

// Default returns a known-good buffer name that can be used to set
// a starting buffer for a given client
//
// c := client.Client(0)
// c.SetBuffer(service.Default)
//
func (s *Service) Default() string {
	return s.buffer
}
