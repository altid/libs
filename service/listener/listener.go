package listener

import (
	"github.com/altid/libs/auth"
	"github.com/altid/libs/service/store"
)

type Listener interface {
	Auth(*auth.Protocol) error
	// Address returns the dialstring of the listening service
	Address() string
	// Connect is a callback for when a client service Connects
	Connect() error
	// Control is a callback for when a control message is sent from a client
	Control() error
	// Listen listens for client connections
	Listen() error
	// Register accepts a Storage, and associates a datatset with the Listener session
	Register(store.Filer) error
}
