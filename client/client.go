package client

import (
	"errors"
	"fmt"
	"io"

	"github.com/lionkov/go9p/p"
)

// MSIZE - maximum size for a message
const MSIZE = p.MSIZE

// CmdType is the type of argument supplied to Ctl
type CmdType int

// CmdTypes here are passed along to Ctl
// The arguments allowed are described in Ctl
const (
	CmdBuffer CmdType = iota
	CmdOpen
	CmdClose
	CmdLink
	CmdRefresh
	CmdConnect
	CmdAttach
	CmdAuth
	CmdTabs
	CmdTitle
	CmdStatus
	CmdAside
	CmdNotify
	CmdInput
	CmdFeed
)

// Client represents a 9p client session
type Client struct {
	run runner
}

// ErrBadArgs is returned from Ctl when incorrect arguments are provided
var ErrBadArgs = errors.New("Too few/incorrect arguments")

type runner interface {
	connect(int) error
	attach() error
	auth() error
	ctl(CmdType, ...string) (int, error) // Just call write at the end in nested types
	tabs() ([]byte, error)
	title() ([]byte, error)
	status() ([]byte, error)
	aside() ([]byte, error)
	input([]byte) (int, error)
	notifications() ([]byte, error)
	feed() (io.ReadCloser, error)
}

// NewClient returns a client ready to connect to addr
func NewClient(addr string) *Client {
	dmc := &client{
		addr: addr,
	}

	return &Client{
		run: dmc,
	}
}

// NewMockClient returns a client for testing
// Feed, if called, will be populated with data from google's GoFuzz every 100ms
func NewMockClient(addr string) *Client {
	dmc := &mock{
		addr:  addr,
		debug: func(CmdType, ...interface{}) {},
	}

	return &Client{
		run: dmc,
	}
}

// Connect performs the network dial for the connection
func (c *Client) Connect(debug int) (err error) {
	return c.run.connect(debug)
}

// Attach is called after optionally calling Auth
func (c *Client) Attach() (err error) {
	return c.run.attach()
}

// Auth is optionally called after Connect to authenticate with the server
func (c *Client) Auth() error {
	return c.run.auth()
}

// Buffer changes the active buffer to the named buffer, or returns an error
func (c *Client) Buffer(name string) (int, error) {
	return c.run.ctl(CmdBuffer, name)
}

// Open attempts to open the named buffer
func (c *Client) Open(name string) (int, error) {
	return c.run.ctl(CmdOpen, name)
}

// Close attempts to close the named buffer
func (c *Client) Close(name string) (int, error) {
	return c.run.ctl(CmdClose, name)
}

// Link updates the current buffer to point to the `to`
func (c *Client) Link(from, to string) (int, error) {
	return c.run.ctl(CmdLink, from, to)
}

// Tabs returns the contents of the `tabs` file for the server
func (c *Client) Tabs() ([]byte, error) {
	return c.run.tabs()
}

// Title returns the contents of the `title` file for a given buffer
func (c *Client) Title() ([]byte, error) {
	return c.run.title()
}

// Status returns the contents of the `status` file for a given buffer
func (c *Client) Status() ([]byte, error) {
	return c.run.status()
}

// Aside returns the contents of the `aside` file for a given buffer
func (c *Client) Aside() ([]byte, error) {
	return c.run.aside()
}

// Input appends the given data string to input
func (c *Client) Input(data []byte) (int, error) {
	return c.run.input(data)
}

// Notifications returns and clears any pending notifications
func (c *Client) Notifications() ([]byte, error) {
	return c.run.notifications()
}

// Feed returns a ReadCloser connected to `feed`. It's expected all reads
// will be read into a buffer with a size of MSIZE
// It is also expected for Feed to be called in its own thread
func (c *Client) Feed() (io.ReadCloser, error) {
	return c.run.feed()
}

func runClientCtl(cmd CmdType, args ...string) ([]byte, error) {
	var data string
	switch cmd {
	case CmdBuffer:
		if len(args) != 1 {
			return nil, ErrBadArgs
		}

		data = fmt.Sprintf("buffer %s\x00", args[0])
	case CmdOpen:
		if len(args) != 1 {
			return nil, ErrBadArgs
		}

		data = fmt.Sprintf("open %s\x00", args[0])
	case CmdClose:
		if len(args) != 1 {
			return nil, ErrBadArgs
		}

		data = fmt.Sprintf("close %s\x00", args[0])
	case CmdLink:
		if len(args) != 2 {
			return nil, ErrBadArgs
		}

		data = fmt.Sprintf("link %s %s\x00", args[0], args[1])
	default:
		return nil, ErrBadArgs
	}

	return []byte(data), nil
}
