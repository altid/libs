package listen9p

import (
	"errors"

	"github.com/altid/libs/auth"
	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/store"
	"github.com/halfwit/styx"
)

type Session struct {
	styx			*styx.Session
	callbacks		callback.Callback
	hasController	bool
	hasConnecter    bool
	address			string
	list			store.Lister
	open			store.Opener
	delete          store.Deleter
}

func NewSession(address string) (*Session, error) {
	return &Session{
		address: address,
		styx: &styx.Session{},
	}, nil
}

// Proxy the auth over the raw connection
func (s *Session) Auth(ap *auth.Protocol) error {
	return nil
}

func (s *Session) Address() string {
	return s.address
}

// Callback when a client connects
func (s *Session) Connect(Username string) error {
	return nil
}

// Callback when a control is issued by clients
func (s *Session) Control() error {
	return nil
}

// Listen on configured network for clients
func (s *Session) Listen() error {
	return nil
}

func (s *Session) Register(filer store.Filer, cbs callback.Callback) error {
	// Verify that we have both functions
	if list, ok := filer.(store.Lister); ok {
		s.list = list
	}

	open, ok := filer.(store.Opener) 
	if !ok {
		return errors.New("Filer does not implement Open")
	}

	s.open = open

	delete, ok := filer.(store.Deleter)
	if ! ok {
		return errors.New("Filer does not implement Delete")
	}

	s.delete = delete

	_, ok = cbs.(callback.Controller)
	if !ok {
		s.hasController = false
	}

	_, ok = cbs.(callback.Connecter)
	if !ok {
		s.hasConnecter = false
	}

	return nil
}

// Technically internal, this is used by Styx
func (s *Session) Serve9P(x *styx.Session) {
	// Callback on client connection
	if s.hasConnecter {
		client := &callback.Client{ 
			Username: x.User,
		}
		s.callbacks.Connect(client)
	}

	for x.Next() {
		req := x.Request()
		f, err := s.open.Open(req.Path())
		switch t := req.(type) {
		case styx.Twalk:
			t.Rwalk(f.Stat())
		case styx.Tstat:
			t.Rstat(f.Stat())
		case styx.Topen:
			t.Ropen(f, err)
// TODO: Handle stream/main, etc
//			switch t.Path() {
//			case "/":
//				t.Ropen(f, err)
//			case "/event":
//				t.Ropen(mkevent(s.User, client))
//			case "/ctrl":
//				t.Ropen(mkctl(fp, s.User, client))
//			case "/tabs":
//				t.Ropen(mktabs(fp, s.User, client))
//			case "/input":
//				t.Ropen(os.OpenFile(fp, os.O_RDWR, 0755))
//			default:
//				t.Ropen(f, err)
//			}
		case styx.Tutimes:
			t.Rutimes(nil)
		case styx.Ttruncate:
			t.Rtruncate(nil)
		// When clients are done with a notification, they delete it. Allow this
		case styx.Tremove:
			switch t.Path() {
			case "/notification", "/notify":
				t.Rremove(s.delete.Delete(req.Path()))
			default:
				t.Rerror("%s", "permission denied")
			}
		}
	}
}

