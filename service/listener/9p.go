package listener

import (
	"errors"

	"github.com/altid/libs/auth"
	"github.com/altid/libs/service/store"
	"github.com/halfwit/styx"
)

// TODO: Somewhere we need a faked network for testing
type Listen9p struct {
	session *styx.Session
	address string
	list	store.Lister
	open    store.Opener
}

// Address returns the address that the service is listening on, as IP:PORT
func (np Listen9p) Address() string {
	return np.address
}

func (np Listen9p) Auth(auth *auth.Protocol) error {
	return nil;
}

// Connect is called when a client connects via 9p
func (np Listen9p) Connect() error {
	return nil
}

// Control is called when a client issues a control command
func (np Listen9p) Control() error {
	return nil
}

// Listen for incoming 9p client connections, handling each in a separate goroutine
func (np Listen9p) Listen() error {
	return styx.ListenAndServe(np.address, np)
}

func (np Listen9p) Register(filer store.Filer) error {
	// Verify that we have both functions
	if list, ok := filer.(store.Lister); ok {
		np.list = list
	}
	
	open, ok := filer.(store.Opener) 
	if !ok {
		return errors.New("Filer does not implement Open")
	}

	np.open = open
	return nil
}

func (np Listen9p) Serve9P(s *styx.Session) {

}
