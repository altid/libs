package listener

import (
	"github.com/altid/libs/auth"
	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/store"
)

// Listener provides a type which can handle incoming client connections
type Listener interface {
	// Auth proxies the auth protocol for authentication
	Auth(*auth.Protocol) error
	// Address returns the dialstring of the listening service
	Address() string
	// Listen listens for client connections
	Listen() error
	// SetActivity is a callback used to notify when a client has read content from a particular buffer
	// It shuold reset the unread count for that given buffer to zero
	SetActivity(func(string))
	// Register accepts a Storage, and associates a Filer datatset with the Listener session
	// Additionally, a callback can be registered to allow on-connect/command information
	Register(store.Filer, commander.Commander, callback.Callback) error
	Type() string
}
