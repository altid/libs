package service

import (
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/controller"
	"github.com/altid/libs/service/internal/control"
)

// We may have to do config rework, but html and markup should be fine.
// We can take the config parsing out to the library
// Just export a function to get that dir --> fd, then pass an fd with handlers for ctl and input messages
// From there, we could convenience wrap our append/create/write/etc

type Handler interface {
	Input([]byte) // add Markup, from buffer
	Ctl(*commander.Command)   // change to command
}

func Start(name string, handler Handler) (controller.Controller, error) {
	ctl, err := control.ConnectService(name)
	if err != nil {
		return nil, err
	}

	go ctl.Listen(handler.Input, handler.Ctl)
	return ctl, nil
}
