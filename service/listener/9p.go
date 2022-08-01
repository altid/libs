package listener

// Export our public interface

import (
	"github.com/altid/libs/auth"
	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/internal/listen9p"
	"github.com/altid/libs/service/store"
)

// Listen9p implements a listener using the 9p protocol
type Listen9p struct {
	session *listen9p.Session
}

func NewListen9p(addr string) *Listen9p {
	session, err := listen9p.NewSession(addr)
	if err != nil {
		return nil
	}

	return &Listen9p{
		session: session,
	}
}

func (np *Listen9p) Auth(ap *auth.Protocol) error {
	return np.session.Auth(ap)
}

func (np *Listen9p) Address() string {
	return np.session.Address()
}

func (np *Listen9p) Listen() error {
	return np.session.Listen()
}

func (np *Listen9p) Register(filer store.Filer, cbs callback.Callback) error {
	return np.session.Register(filer, cbs) 
}
