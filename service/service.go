package service

import (
	"log"
	"os"
	"sort"

	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/control"
	"github.com/altid/libs/service/internal/command"
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
	callback  callback.Callback
	commander commander.Commander
	listener  listener.Listener
	runner    runner.Runner
	store     store.Filer

	cmdlist []*commander.Command
	control *control.Control
	name    string
	debug   func(serviceMsg, ...interface{})
}

func New(name string, debug bool) *Service {
	// Service will get store/listener/callback/runner
	s := &Service{
		name: name,
	}

	if debug {
		s.debug = serviceLogger
	}

	return s
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
	var command command.Command

	// Internal:
	// set up store
	// set up listener
	// register callbacks
	// register runner - if no runner, fail!
	// start control listens

	s.commander = &command
	s.debug(serviceCommand)

	// Add defaults
	s.cmdlist = commander.DefaultCommands
	sort.Sort(commander.CmdList(s.cmdlist))
	s.debug(serviceSetCommands, s.cmdlist)

	// Finally run
	s.debug(serviceStarted)
	return nil
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
		if _, ok := args[0].(callback.Controller); ok {
			l.Println("callback: control message callback registered")
		}
	case serviceSetCommands:
		for _, arg := range args {
			if cmd, ok := arg.(commander.Command); ok {
				l.Printf("adding command: %v", cmd.Name)
			}
		}
	case serviceCommand:
		l.Println("commander registered")
	case serviceControl:
		l.Println("control registered")
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
