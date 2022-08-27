package callback

import (
	"github.com/altid/libs/service/controller"
)

type Connecter interface {
	Connect(*Client, controller.Controller) error
}

type Callback interface {
	Connecter
}

// Sender interface is used by the listeners to handle control messages
// SendCommand can be intercepted, but finally should call your service.SendCommand from your controller with the payload
type Sender interface {
	SendCommand(string) error
}

// A client is returned on Client connection
type Client struct {
	Username string
}
