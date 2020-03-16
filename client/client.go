package client

import (
	"errors"
	"fmt"
	"io"

	"github.com/altid/libs/fs"
	"github.com/lionkov/go9p/p"
)

// MSIZE - maximum size for a message
const MSIZE = p.MSIZE

// Used internally
const (
	CmdBuffer int = iota
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
	CmdDocument
	CmdComm
)

// Client represents a 9p client session
type Client struct {
	run runner
}

// ErrBadArgs is returned from Ctl when incorrect arguments are provided
var ErrBadArgs = errors.New("Too few/incorrect arguments")

type runner interface {
	cleanup()
	connect(int) error
	attach() error
	auth() error
	command(*fs.Command) error
	ctl(int, ...string) (int, error)
	tabs() ([]byte, error)
	title() ([]byte, error)
	status() ([]byte, error)
	aside() ([]byte, error)
	input([]byte) (int, error)
	notifications() ([]byte, error)
	feed() (io.ReadCloser, error)
	document() ([]byte, error)
	getCommands() ([]*fs.Command, error)
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
		debug: func(int, ...interface{}) {},
	}

	return &Client{
		run: dmc,
	}
}

// GetCommands returns a list of available commands for the connected service
func (c Client) GetCommands() ([]*fs.Command, error) {
	return c.run.getCommands()
}

// Document returns the contents of a document file on the host
// if it exists, or an error
func (c *Client) Document() ([]byte, error) {
	return c.run.document()
}

// Cleanup closes the underlying connection
func (c *Client) Cleanup() {
	c.run.cleanup()
}

// Connect performs the network dial for the connection
func (c *Client) Connect(debug int) (err error) {
	return c.run.connect(debug)
}

// Command sends the named command to the service
// If command is invalid, it will return an error
func (c *Client) Command(cmd *fs.Command) error {
	return c.run.command(cmd)
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

// FeedIterator allows you to step through lines of feed with Next()
// Useful for gomobile, etc
type FeedIterator struct {
	rc io.ReadCloser
}

// FeedIterator returns a new FeedIterator ready to go
func (c *Client) FeedIterator() (*FeedIterator, error) {
	f, err := c.run.feed()
	if err != nil {
		return nil, err
	}

	return &FeedIterator{f}, nil
}

// Next will return the next slice of bytes, or an error
// After an error, future calls to Next() will panic
func (f *FeedIterator) Next() ([]byte, error) {
	b := make([]byte, MSIZE)
	if _, err := f.rc.Read(b); err != nil {
		defer f.rc.Close()
		return nil, err
	}

	return b, nil
}

func runClientCtl(cmd int, args ...string) ([]byte, error) {
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
