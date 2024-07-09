package service

import (
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/controller"
	"github.com/altid/libs/service/internal/control"
)

func Start(name string, cmd func(*commander.Command)) (controller.Controller, error) {
	ctl, err := control.ConnectService(name)
	if err != nil {
		return nil, err
	}

	go ctl.Listen(cmd)
	return ctl, nil
}
