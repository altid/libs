package runner

import (
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/controller"
)

// Starter must not block when called, and return immediately with any errors
type Starter interface {
	Start(controller.Controller) error
}

// Listen must block, or the Service will close
type Listener interface {
	Listen(controller.Controller)
}

// A Runner represents the underlying service
// On commands issued via the listener, Command() will be called with the payload
// Any errors encountered should be returned, and will be passed to the connecting client
// Runner must implement either a runner.Listener or a runner.Starter as well as Command
type Runner interface {
	Command(*commander.Command) error
}
