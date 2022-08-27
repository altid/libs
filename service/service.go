package service

import (
	"context"
	"log"
	"os"
	"sort"

	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/controller"
	"github.com/altid/libs/service/internal/session"
	"github.com/altid/libs/service/listener"
	"github.com/altid/libs/service/runner"
	"github.com/altid/libs/store"
)

type serviceMsg int

const (
	serviceStore serviceMsg = iota
	serviceListener
	serviceCallbacks
	serviceRunner
	serviceControl
	serviceCommand
	serviceSetCommands
	serviceStarted
	serviceError
)

type Service struct {
	ctx      context.Context
	callback callback.Callback
	control  controller.Controller
	listener listener.Listener
	runner   runner.Runner
	store    store.Filer

	cmdlist []*commander.Command
	name    string
	address string
	debug   func(serviceMsg, ...interface{})
}

func New(name string, address string, debug bool) *Service {
	s := &Service{
		name:    name,
		address: address,
	}

	if debug {
		s.debug = serviceLogger
	}

	return s
}

func (s *Service) WithContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *Service) WithStore(st store.Filer) {
	s.debug(serviceStore, st.Type())
	s.store = st
}

func (s *Service) WithListener(l listener.Listener) {
	s.debug(serviceListener, l.Type())
	s.listener = l
}

func (s *Service) WithCallbacks(cb callback.Callback) {
	s.debug(serviceCallbacks, cb)
	s.callback = cb
}

func (s *Service) WithRunner(r runner.Runner) {
	s.debug(serviceRunner, r)
	s.runner = r
}

func (s *Service) SetCommands(cmds []*commander.Command) {
	s.debug(serviceSetCommands, cmds)
	s.cmdlist = append(s.cmdlist, cmds...)
	sort.Sort(commander.CmdList(s.cmdlist))
}

func (s *Service) Listen() error {
	session := &session.Session{
		Ctx:      s.ctx,
		Callback: s.callback,
		Control:  s.control,
		Listener: s.listener,
		Runner:   s.runner,
		Store:    s.store,

		Name:    s.name,
		Address: s.address,
	}

	if s.debug == nil {
		return session.Listen(false)
	}

	return session.Listen(true)
}

// Very good logging is beneficial!
func serviceLogger(msg serviceMsg, args ...interface{}) {
	l := log.New(os.Stdout, "service ", 0)

	switch msg {
	case serviceError:
		l.Printf("error: loc=\"%s\" err=\"%v\"", args[0], args[1])
	case serviceCallbacks:
		if _, ok := args[0].(callback.Connecter); ok {
			l.Println("callback: client connection callback registered")
		}
	case serviceSetCommands:
		for _, arg := range args {
			if cmd, ok := arg.(commander.Command); ok {
				l.Printf("adding command: %v", cmd.Name)
			}
		}
	case serviceListener:
		l.Printf("listener: type=\"%s\"", args[0])
	case serviceRunner:
		if _, ok := args[0].(runner.Listener); ok {
			l.Println("runner: registered Listener")
		}
		if _, ok := args[0].(runner.Starter); ok {
			l.Println("runner: registered Starter")
		}
	case serviceStore:
		l.Printf("store: type=\"%s\"", args[0])
	case serviceStarted:
		l.Println("started")
	}
}
