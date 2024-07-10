package callback

import (
	"github.com/altid/libs/markup"
	"github.com/altid/libs/service/controller"
)

type Connecter interface {
	Connect(Username string) error
}

type Callback interface {
	Connecter
	Handler
	Starter
}

// Handler is called when data is written to an `input` file
// The path will be the buffer in which the data was written
type Handler interface {
	Handle(path string, c *markup.Lexer) error
}

// Sender interface is used by the listeners to handle control messages
// SendCommand can be intercepted, but finally should call your service.SendCommand from your controller with the payload
type Sender interface {
	SendCommand(string) error
}

// Starter is called to start the main loop of the client
type Starter interface {
	Start(controller.Controller) error
}