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
// Services will register either Start or Listen in their code, which acts as an entrypoint
// On commands issued via the listener, Command() will be called with the payload
// Any errors encountered should be returned, and will be passed to the connecting client
type Runner interface {
	Listener
	Starter
	Command(*commander.Command) error
}
