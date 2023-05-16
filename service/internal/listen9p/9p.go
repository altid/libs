package listen9p

import (
	"io/fs"
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
	sessionInfo
	sessionErr
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
	debug   func(sessionMsg, ...any)
}

func NewSession(address string, key, cert string, debug bool) (*Session, error) {
	s := &Session{
		address: address,
		styx:    &styx.Session{},
		key:     key,
		cert:    cert,
		debug:   func(sessionMsg, ...any) {},
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

	if delete, ok := filer.(store.Deleter); ok {
		s.delete = delete
	}

	s.cmd = cmd
	s.cb = cb

	return nil
}

// Technically internal, this is used by Styx
func (s *Session) Serve9P(x *styx.Session) {
	client := &client{
		s:       s,
		uuid:    uuid.New(),
		name:    x.User,
	}
	// Somewhat obtuse method of discovering a valid buffername
	// Walk through the list of files, and check if it as top level
	// We could also check against our known list of files such as /ctrl
	l := s.list.List()
	for _, c := range l {
		if path.Dir(c) != "/" && path.Dir(c) != "." {
			client.current = path.Dir(c)
			break
		}
	}
	s.debug(sessionInfo, "current", client.current)
	if x.Access != "" {
		client.current = x.Access
	}

	if e := s.cb.Connect(x.User); e != nil {
		s.debug(sessionErr, e)
		return
	}

	s.debug(sessionClient, client)
	for x.Next() {
		req := x.Request()
		switch t := req.(type) {
		case styx.Twalk:
			t.Rwalk(client.walk(req.Path()))
		case styx.Tstat:
			t.Rstat(client.stat(req.Path()))
		case styx.Topen:
			if t.Path() == "/" {
				t.Ropen(client.s.open.Root(req.Path()))
			} else {
				// We should reopen here
				t.Ropen(client.open(req.Path()))
			}
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

type client struct {
	s       *Session
	name    string
	uuid    uuid.UUID
	closer  func() error
	current string
}

func (c *client) stat(buffer string) (fs.FileInfo, error) {
	if buffer == "/" {
		r, err := c.s.open.Root(buffer)
		if err != nil {
			return nil, err
		}
		return r.Info()
	}
	f, err := c.getFile(buffer)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	return f.Stat()
}

func (c *client) walk(buffer string) (fs.FileInfo, error) {
	return c.stat(buffer)
}

func (c *client) open(buffer string) (store.File, error) {
	f, err := c.getFile(buffer)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// Callback passed to ctrl's open, we get this on write
// we want to intercept buffer commands
func (c *client) ctrlWrite(ctrl []byte) error {
	cmd, err := c.s.cmd.FromBytes(ctrl)
	if err != nil {
		return nil
	}

	switch cmd.Name {
	case "buffer":
		c.s.debug(sessionBuffer, cmd)
		if c.closer != nil {
			c.closer()
		}
		c.current = cmd.Args[0]
		return nil
	case "open", "link":
		// If we have a feed going, close it here
		if c.closer != nil {
			c.closer()
		}
	}

	// This doesn't seem right
	r := c.s.cmd.RunCommand()
	return r(cmd)
}

func (c *client) ctrlData() []byte {
	r := c.s.cmd.CtrlData()
	return r()
}

// Build the files from the store, do not produce as-is since they'll be broken
func (c *client) getFile(in string) (store.File, error) {
	switch in {
	case "/errors":
		return c.s.open.Open("/errors")
	case "/tabs":
		return c.s.open.Open("/tabs")
	case "/ctrl":
		return files.Ctrl(c.ctrlWrite, c.ctrlData)
	case "/input":
		return files.Input(c.current, c.s.cb)
	case "/feed":
		fp := path.Join("/", c.current, in)
		f, err := c.s.open.Open(fp)
		// We need to signal the store that we're done with the Reads on Feed
		// Send a close when we change current buffer/open/link
		c.closer = f.Close
		return files.Feed(f, err)
	default:
		fp := path.Join("/", c.current, in)
		for _, item := range c.s.list.List() {
			if item == fp {
				return c.s.open.Open(fp)
			}
		}
	}
	return nil, Err9pFileNotFound
}

func sessionLogger(msg sessionMsg, args ...any) {
	switch msg {
	case sessionErr:
		l.Printf("error: %e", args[0])
	case sessionInfo:
		l.Printf("info: %v", args)
	case sessionStart:
		l.Println("starting session")
	case sessionOpen:
		l.Printf("open: %s", args[0])
	case sessionBuffer:
		if cmd, ok := args[0].(*commander.Command); ok {
			l.Printf("buffer: name=\"%s\" args=\"%s\" from=\"%s\"", cmd.Name, cmd.Args[0], cmd.From)
		}
	case sessionClient:
		if client, ok := args[0].(*client); ok {
			l.Printf("client: user=\"%s\" buffer=\"%s\" id=\"%s\"\n", client.name, client.current, client.uuid.String())
		}
	}
}
