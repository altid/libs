package listen9p

import (
	"log"
	"os"
	"path"

	"github.com/altid/libs/auth"
	"github.com/altid/libs/service/callback"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/internal/files"
	"github.com/altid/libs/store"
	"github.com/google/uuid"
	"github.com/halfwit/styx"
)

var l *log.Logger

type Err9p string

const (
	Err9pNoOpen       = Err9p("store does not implement Opener")
	Err9pFileNotFound = Err9p("file not found")
)

type sessionMsg int

const (
	sessionStart sessionMsg = iota
	sessionClient
	sessionBuffer
	sessionOpen
)

type Session struct {
	styx    *styx.Session
	key     string
	cert    string
	address string
	cmd     commander.Commander
	cb      callback.Callback
	list    store.Lister
	open    store.Opener
	delete  store.Deleter
	stream  store.Streamer
	debug   func(sessionMsg, ...interface{})
}

func NewSession(address string, key, cert string, debug bool) (*Session, error) {
	s := &Session{
		address: address,
		styx:    &styx.Session{},
		key:     key,
		cert:    cert,
		debug:   func(sessionMsg, ...interface{}) {},
	}

	if debug {
		s.debug = sessionLogger
		l = log.New(os.Stdout, "listen9p ", 0)
	}

	return s, nil
}

func (e Err9p) Error() string {
	return string(e)
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
func (s *Session) Register(filer store.Filer, cmd commander.Commander, cb callback.Callback) error {
	if list, ok := filer.(store.Lister); ok {
		s.list = list
	}
	open, ok := filer.(store.Opener)
	if !ok {
		return Err9pNoOpen
	}
	s.open = open

	if stream, ok := filer.(store.Streamer); ok {
		s.stream = stream
	}

	if delete, ok := filer.(store.Deleter); ok {
		s.delete = delete
	}

	s.cmd = cmd
	s.cb = cb

	return nil
}

// Build the files from the store, do not produce as-is since they'll be broken
func getFile(c *Client, in string) (store.File, error) {
	switch in {
	case "/":
		// Returns our base dir with our current buffer keyed
		return c.s.open.Root(c.current)
	case "/errors":
		return c.s.open.Open("/errors")
	case "/tabs":
		return c.s.open.Open("/tabs")
		// do a tabs thing
	case "/ctrl":
		// Allow buffer modifications
		return files.Ctrl(c.ctrlWrite, c.ctrlData)
	case "/input":
		return files.Input(c.current, c.s.cb)
	case "/feed":
		feed, err := c.s.open.Open(path.Join(c.current, "feed"))
		if err != nil {
			return nil, err
		}

		return files.Feed(c.current, c.s.stream, feed)
	default:
		fp := path.Join(c.current, in)
		for _, item := range c.s.list.List() {
			// Only open items we have buffers open with
			if path.Clean(item) == c.current {
				c.s.debug(sessionOpen, fp)
				return c.s.open.Open(fp)
			}
		}
	}
	return nil, Err9pFileNotFound
}

// Technically internal, this is used by Styx
func (s *Session) Serve9P(x *styx.Session) {
	client := &Client{
		s:       s,
		uuid:    uuid.New(),
		name:    x.User,
		current: "server",
	}

	s.debug(sessionClient, client)

	files := make(map[string]store.File)

	for x.Next() {
		req := x.Request()
		f, err := getFile(client, req.Path())
		if err != nil {
			req.Rerror("%s", err)
			continue
		}

		log.Printf("Looping with current: %s\n", client.current)
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

type Client struct {
	s       *Session
	name    string
	uuid    uuid.UUID
	current string
}

// Callback passed to ctrl's open, we get this on write
// we want to intercept buffer commands
func (c *Client) ctrlWrite(ctrl []byte) error {
	cmd, err := c.s.cmd.FromBytes(ctrl)
	if err != nil {
		return nil
	}

	switch cmd.Name {
	case "buffer":
		c.s.debug(sessionBuffer, cmd)
		c.current = cmd.Args[0]
		return nil
	}

	// This doesn't seem right
	r := c.s.cmd.RunCommand()
	return r(cmd)
}

func (c *Client) ctrlData() []byte {
	r := c.s.cmd.CtrlData()
	return r()
}

func sessionLogger(msg sessionMsg, args ...interface{}) {
	switch msg {
	case sessionStart:
		l.Println("starting session")
	case sessionOpen:
		l.Printf("open: %s", args[0])
	case sessionBuffer:
		if cmd, ok := args[0].(*commander.Command); ok {
			l.Printf("buffer: name=\"%s\" args=\"%s\" from=\"%s\"", cmd.Name, cmd.Args[0], cmd.From)
		}
	case sessionClient:
		if client, ok := args[0].(*Client); ok {
			l.Printf("client: user=\"%s\" buffer=\"%s\" id=\"%s\"\n", client.name, client.current, client.uuid.String())
		}
	}
}
