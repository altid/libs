package listener

import (
	"log"
	"os"

	"github.com/altid/libs/auth"
	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/internal/listenssh"
	"github.com/altid/libs/store"
)

type NewSsh int

type ListenSSH struct {
	session *listenssh.Session
	debug   func(listenMsg, ...any)
}

// NewListenSSH returns a listener that will provide an interactive SSH endpoint
// A client can dial in, such as `ssh myservice.net` to interact with an Altid service
// Address is the address for the SSH server to bind to
func NewListenSSH(addr, id_rsa string, debug bool) (ListenSSH, error) {
	session, err := listenssh.NewSession(addr, id_rsa, debug)
	ls := ListenSSH{
		session: session,
	}
	if debug {
		l = log.New(os.Stdout, "ssh ", 0)
		ls.debug = listenLogger
	}
	return ls, err
}

func (ssh ListenSSH) Auth(ap *auth.Protocol) error {
	ssh.debug(listenAuth, ap)
	return ssh.session.Auth(ap)
}

func (ssh ListenSSH) Address() string {
	ssh.debug(listenAddress, ssh.session.Address())
	return ssh.session.Address()
}

func (ssh ListenSSH) Register(filer store.Filer, cmd commander.Commander, cb callback.Callback) error {
	return ssh.session.Register(filer, cmd, cb)
}

func (ssh ListenSSH) Listen() error                { return ssh.session.Listen() }
func (ssh ListenSSH) SetActivity(act func(string)) { ssh.session.SetActivity(act) }
func (ssh ListenSSH) Type() string                 { return "ssh" }
