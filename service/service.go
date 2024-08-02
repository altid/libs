package service

import (
	"context"

	"github.com/altid/libs/threads"
	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/internal/control"
)

type Service struct {
	fg  bool
	ctx context.Context
	name string
	cb callback.Callback
	cmds []*commander.Command
	ctl *control.Control
}

func Register(ctx context.Context, name string, fg bool) (*Service, error) {
	return &Service{
		ctx: ctx,
		name: name,
		fg: fg,
	}, nil
}

func (s *Service) SetCommands(cmds []*commander.Command) {
	s.cmds = cmds
}

func (s *Service) SetCallbacks(cb callback.Callback) {
	s.cb = cb
}

func (s *Service) Listen() error {
	return threads.Start(func () error{
		// Make sure we call everything after the fork to set up our stack
		ctl, err := control.ConnectService(s.ctx, s.name)
		if err != nil {
			return err
		}
		ctl.SetCommands(s.cmds)
		ctl.SetCallbacks(s.cb)
		s.ctl = ctl
		return s.ctl.Listen()
	}, s.fg)
	//return threads.Start(s.ctl.Listen, s.fg)
}
