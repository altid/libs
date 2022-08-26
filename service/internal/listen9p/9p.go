package listen9p

import (
	"errors"
	"path"

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
	key				string
	cert 			string
	address			string
	list			store.Lister
	open			store.Opener
	delete          store.Deleter
}

func NewSession(address string, key, cert string) (*Session, error) {
	s := &Session{
		address: address,
		styx: &styx.Session{},
		key: key,
		cert: cert,
	}

	return s, nil
}

// Proxy the auth over the raw connection
func (s *Session) Auth(ap *auth.Protocol) error {
	return nil
}

func (s *Session) Address() string {
	return s.address
}

// Listen on configured network for clients
func (s *Session) Listen() error {
	if s.key == "none" || s.key == "" {
		if s.cert == "none" || s.cert == "" {
			return styx.ListenAndServe(s.address, s)
		}
	}

	return styx.ListenAndServeTLS(s.address, s.key, s.cert, s)

}

func (s *Session) Register(filer store.Filer, cbs callback.Callback) error {
	if list, ok := filer.(store.Lister); ok {
		s.list = list
	}

	open, ok := filer.(store.Opener)
	if ! ok {
		return errors.New("store does not implement required 'Open'")
	}
	s.open = open

	if delete, ok := filer.(store.Deleter); ok {
		s.delete = delete
	}

	if _, ok := cbs.(callback.Controller); ! ok {
		s.hasController = false
	}

	if _, ok := cbs.(callback.Connecter); !ok {
		s.hasConnecter = false
	}

	return nil
}

// Build the files from the store, do not produce as-is since they'll be broken
func getFile(s *Session, current, in string) (store.File, error) {
	switch(in) {
	case "/":
		// Returns our base dir with our current buffer keyed
		return s.open.Root(current)
	case "/errors":
		return s.open.Open("/errors")
	case "/tabs":
		return nil, nil
		// do a tabs thing
	case "/ctrl":
		return nil, nil
		// do a special ctrl thing
	case "/input":
		return nil, nil
		// do a special input thing
	default:
		fp := path.Join(current, in)
		return s.open.Open(fp)
	}
}

// Technically internal, this is used by Styx
func (s *Session) Serve9P(x *styx.Session) {
	var current string

	// Callback on client connection
	if s.hasConnecter {
		client := &callback.Client{ 
			Username: x.User,
		}
		s.callbacks.Connect(client)
	}

	files := make(map[string]store.File)
	current = "server"

	for x.Next() {
		req := x.Request()
		f, err := getFile(s, current, req.Path())
		if err != nil {
			req.Rerror("%s", err)
		}

		switch t := req.(type) {
		case styx.Twalk:
			t.Rwalk(f.Stat())
		case styx.Tstat:
			t.Rstat(f.Stat())
		case styx.Topen:
			t.Ropen(f, nil)
		case styx.Tutimes:
			t.Rutimes(nil)
		case styx.Ttruncate:
			t.Rtruncate(nil)
		// When clients are done with a notification, they delete it. Allow this
		case styx.Tremove:
			switch t.Path() {
			case "/notification", "/notify":
				delete(files, f.Name())
				t.Rremove(s.delete.Delete(req.Path()))
			default:
				t.Rerror("%s", "permission denied")
			}
		}
	}
}

