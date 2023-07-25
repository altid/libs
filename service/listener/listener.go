package listener

import (
	"log"

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
	// This is largely implementation specific how this interface will be satisfied
	Listen() error
	// SetActivity is a callback used to notify when a client has read content from a particular buffer
	// It should reset the unread count for that given buffer to zero
	SetActivity(func(string))
	// Register accepts a Storage, and associates a Filer datatset with the Listener session
	// Additionally, a callback can be registered to allow on-connect/command information
	Register(store.Filer, commander.Commander, callback.Callback) error
	// Type returns a string representation of the name of the listener
	Type() string
}

var l *log.Logger

type listenMsg int

const (
	listenAuth listenMsg = iota
	listenAddress
	listenListen
	listenRegister
)

func listenLogger(msg listenMsg, args ...any) {
	switch msg {
	case listenAuth:
		//auth := args[0].(*auth.Protocol)
		//l.Printf("auth: \"%s\"", auth.Info())
		l.Printf("auth: %s", args[0])
	case listenAddress:
		l.Printf("address: \"%s\"", args[0])
	case listenListen:
		l.Println("starting")
	case listenRegister:
		l.Printf("register: filer=\"%s\" callbacks=\"%s\" cmd=\"%s\"", args[0], args[1], args[2])
	}
}
