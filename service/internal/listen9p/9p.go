package listen9p

import (
	"errors"
	"path"

	"github.com/altid/libs/auth"
	"github.com/altid/libs/service/internal/files"
	"github.com/altid/libs/store"
	"github.com/halfwit/styx"
)

type Session struct {
	styx    *styx.Session
	key     string
	cert    string
	address string
	current string
	list    store.Lister
	open    store.Opener
	delete  store.Deleter
}

func NewSession(address string, key, cert string) (*Session, error) {
	s := &Session{
		address: address,
		styx:    &styx.Session{},
		key:     key,
		cert:    cert,
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

// Here we need a command channel as well to move it out of listen
// Wrap in a type so we can not do bare channels
func (s *Session) Register(filer store.Filer) error {
	if list, ok := filer.(store.Lister); ok {
		s.list = list
	}
	open, ok := filer.(store.Opener)
	if !ok {
		return errors.New("store does not implement required 'Open'")
	}
	s.open = open

	if delete, ok := filer.(store.Deleter); ok {
		s.delete = delete
	}

	return nil
}

// Build the files from the store, do not produce as-is since they'll be broken
func getFile(s *Session, current, in string) (store.File, error) {
	switch in {
	case "/":
		// Returns our base dir with our current buffer keyed
		return s.open.Root(current)
	case "/errors":
		return s.open.Open("/errors")
	case "/tabs":
		return s.open.Open("/tabs")
		// do a tabs thing
	case "/ctrl":
		// Allow buffer modifications
		d, err := files.Ctrl()
		if err != nil {
			return nil, err
		}
		go func(s *Session, ctrl *files.CtrlFile) {
			if current, ok := <-ctrl.Current; ok {
				s.current = current
			}
		}(s, d)
		return d, nil
	case "/input":
		return s.open.Open("/input")
		// do a special input thing
	default:
		fp := path.Join(current, in)
		for _, item := range s.list.List() {
			// Only open items we have buffers open with
			if path.Clean(item) == current {
				return s.open.Open(fp)
			}
		}
	}
	return nil, errors.New("file not found")
}

// Technically internal, this is used by Styx
func (s *Session) Serve9P(x *styx.Session) {
	var current string

	files := make(map[string]store.File)
	s.current = "server"

	for x.Next() {
		req := x.Request()
		f, err := getFile(s, current, req.Path())
		if err != nil {
			req.Rerror("%s", err)
			continue
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
