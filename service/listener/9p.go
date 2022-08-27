package listener

// Export our public interface

import (
	"log"
	"os"

	"github.com/altid/libs/auth"
	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/internal/listen9p"
	"github.com/altid/libs/store"
)

type listenMsg int

const (
	listenAuth listenMsg = iota
	listenAddress
	listenListen
	listenRegister
)

// Listen9p implements a listener using the 9p protocol
type Listen9p struct {
	session *listen9p.Session
	debug   func(listenMsg, ...interface{})
}

// NewListen9p returns a new listener
// If a key and cert are provided, the listener will use TLS
func NewListen9p(addr string, key, cert string, debug bool) (Listen9p, error) {
	session, err := listen9p.NewSession(addr, key, cert)
	l := Listen9p{
		session: session,
	}

	if debug {
		l.debug = listenLogger
	}

	return l, err
}

func (np Listen9p) Auth(ap *auth.Protocol) error {
	np.debug(listenAuth, ap)
	return np.session.Auth(ap)
}

func (np Listen9p) Address() string {
	np.debug(listenAddress, np.session.Address())
	return np.session.Address()
}

func (np Listen9p) Listen() error {
	return np.session.Listen()
}

// TODO: This almost certainly changes, but we shall see
func (np Listen9p) Register(filer store.Filer, cbs callback.Callback, cmd callback.Sender) error {
	return np.session.Register(filer, cbs, cmd) 
}

func (np Listen9p) Type() string {
	return "9p"
}

func listenLogger(msg listenMsg, args ...interface{}) {
	l := log.New(os.Stdout, "9p ", 0)

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