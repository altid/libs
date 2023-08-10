package session

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"sort"

	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/controller"
	"github.com/altid/libs/service/internal/command"
	"github.com/altid/libs/service/internal/files"
	"github.com/altid/libs/service/listener"
	"github.com/altid/libs/service/runner"
	"github.com/altid/libs/store"
)

var l *log.Logger

type sessionMsg int

const (
	sessionError sessionMsg = iota
	sessionStart
	sessionSetCommands
	sessionDefaultStore
)

type Session struct {
	Ctx       context.Context
	Callback  callback.Callback
	Control   controller.Controller
	Listeners []listener.Listener
	Runner    runner.Runner
	Store     store.Filer
	Sender    callback.Sender
	commander commander.Commander
	cmdlist   []*commander.Command
	Name      string
	Address   string
	debug     func(sessionMsg, ...any)
}

func (s *Session) Listen(debug bool) error {
	if debug {
		s.debug = sessionLogger
		l = log.New(os.Stdout, "session ", 0)
	}
	s.debug(sessionStart)
	if s.Store == nil {
		s.debug(sessionDefaultStore)
		s.Store = store.NewRamstore(debug)
	}
	filer := files.New(s.Store, debug)
	s.Control = filer
	s.commander = &command.Command{
		SendCommand:     s.sendCommand,
		CtrlDataCommand: s.ctrlData,
	}
	s.cmdlist = commander.DefaultCommands
	sort.Sort(commander.CmdList(s.cmdlist))
	s.debug(sessionSetCommands, s.cmdlist)
	// If we have no listeners, set up a barebones TCP listener over 9p
	// Otherwise just set up like normal
	if len(s.Listeners) == 0 {
		listener, err := listener.NewListen9p(s.Address, "", "", debug)
		if err != nil {
			return err
		}
		s.Listeners = append(s.Listeners, listener)
	}
	// Listen on all, send in the chans here
	listening := false
	echan := make(chan error)
	for _, li := range s.Listeners {
		if e := li.Register(s.Store, s.commander, s.Callback); e != nil {
			return e
		}
		li.SetActivity(filer.Activity)
		listening = true
		go func(echan chan error, li listener.Listener) {
			echan <- li.Listen()
		}(echan, li)
	}
	if !listening {
		return errors.New("unable to start valid listener")
	}
	// These are idempotent in that the service will start
	// but offer some flexibility in the event loops
	if svc, ok := s.Runner.(runner.Listener); ok {
		go svc.Listen(s.Control)
	} else if svc, ok := s.Runner.(runner.Starter); ok {
		if e := svc.Start(s.Control); e != nil {
			return e
		}
	} else {
		err := errors.New("invalid/nil runner supplied")
		return err
	}
	// We could spin up an echan for each instance, or refcount and error only on absolute failure of all listeners
	// but a listener failing will result in undesired behaviour regardless versus the configurations
	select {
	case <-s.Ctx.Done():
		return nil
	case err := <-echan:
		return err
	}
}

func (s *Session) sendCommand(cmd *commander.Command) error {
	switch cmd.Name {
	case "shutdown":
		s.Ctx.Done()
	case "reload":
	case "restart":
	}
	return s.Runner.Command(cmd)
}

func (s *Session) ctrlData() (b []byte) {
	cw := bytes.NewBuffer(b)
	s.commander.WriteCommands(s.cmdlist, cw)
	return cw.Bytes()
}

func sessionLogger(msg sessionMsg, args ...any) {
	switch msg {
	case sessionError:
		l.Printf("error: %v", args[0])
	case sessionStart:
		l.Println("started")
	case sessionDefaultStore:
		l.Println("using default store")
	case sessionSetCommands:
		for _, arg := range args {
			if cmd, ok := arg.(commander.Command); ok {
				l.Printf("adding command: %v", cmd.Name)
			}
		}
	}
}
