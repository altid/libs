package service

import (
	"context"

	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/internal/control"
)

type Service struct {
	ctl *control.Control
}

func Register(ctx context.Context, name string) (*Service, error) {
	ctl, err := control.ConnectService(ctx, name)
	if err != nil {
		return nil, err
	}

	return &Service{
		ctl: ctl,
	}, nil
}

func (s *Service) SetCommands(cmds []*commander.Command) {
	s.ctl.SetCommands(cmds)
}

func (s *Service) SetCallbacks(cb callback.Callback) {
	s.ctl.SetCallbacks(cb)
}

func (s *Service) Listen() error {
	return s.ctl.Listen()
}
