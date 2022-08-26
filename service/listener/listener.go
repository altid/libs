package listener

import (
	"github.com/altid/libs/auth"
	"github.com/altid/libs/store"
	"github.com/altid/libs/service/callback"
)

// Listener provides a type which can handle incoming client connections
type Listener interface {
	// Auth proxies the auth protocol for authentication
	Auth(*auth.Protocol) error
	// Address returns the dialstring of the listening service
	Address() string
	// Listen listens for client connections
	Listen() error
	// Register accepts a Storage, and associates a Filer datatset with the Listener session
	// Additionally, a callback can be registered to allow on-connect/command information
	Register(store.Filer, callback.Callback) error
}
