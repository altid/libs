package listener

// Export our public interface

import (
	"log"
	"os"

	"github.com/altid/libs/auth"
	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/internal/listen9p"
	"github.com/altid/libs/store"
)

// Listen9p implements a listener using the 9p protocol
type Listen9p struct {
	session *listen9p.Session
	debug   func(listenMsg, ...any)
}

// NewListen9p returns a new listener
// If a key and cert are provided, the listener will use TLS
func NewListen9p(addr string, key, cert string, debug bool) (Listen9p, error) {
	session, err := listen9p.NewSession(addr, key, cert, debug)
	lp := Listen9p{
		session: session,
	}
	if debug {
		l = log.New(os.Stdout, "9p ", 0)
		lp.debug = listenLogger
	}
	return lp, err
}

func (np Listen9p) Auth(ap *auth.Protocol) error {
	np.debug(listenAuth, ap)
	return np.session.Auth(ap)
}

func (np Listen9p) Address() string {
	np.debug(listenAddress, np.session.Address())
	return np.session.Address()
}

func (np Listen9p) Register(filer store.Filer, cmd commander.Commander, cb callback.Callback) error {
	return np.session.Register(filer, cmd, cb)
}

func (np Listen9p) SetActivity(act func(string)) { np.session.SetActivity(act) }
func (np Listen9p) Listen() error                { return np.session.Listen() }
func (np Listen9p) Type() string                 { return "9p" }
