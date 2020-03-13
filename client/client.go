package client

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/knieriem/g/go9p/user"
	"github.com/lionkov/go9p/p"
	"github.com/lionkov/go9p/p/clnt"
)

// Track fids instead in a map as we use them

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
)

// ErrBadArgs is returned from Ctl when incorrect arguments are provided
var ErrBadArgs = errors.New("Too few/incorrect arguments")

// Client represents a 9p client session
type Client struct {
	afid   *clnt.Fid
	root   *clnt.Fid
	addr   string
	buffer string
	clnt   *clnt.Clnt
}

// NewClient returns an authenticated client
func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

// Connect performs the network dial for the connection
func (c *Client) Connect(debug int) (err error) {
	dial := fmt.Sprintf("%s:564", c.addr)

	conn, err := net.Dial("tcp", dial)
	if err != nil {
		return err
	}

	c.clnt, err = clnt.Connect(conn, p.MSIZE, false)
	if err != nil {
		return err
	}

	c.clnt.Debuglevel = debug

	return
}

// Attach is called after optionally calling Auth
func (c *Client) Attach() (err error) {
	root, err := c.clnt.Attach(c.afid, user.Current(), "/")
	if err != nil {
		return err
	}

	c.root = root

	return nil
}

// Auth is optionally called after Connect to authenticate with the server
func (c *Client) Auth() error {
	afid, err := c.clnt.Auth(user.Current(), "/")
	if err != nil {
		return err
	}

	c.afid = afid
	return nil
}

// Ctl sends the given arguments and named command to the connected Ctl file
// The arguments expected differ for each type, and will result in an error
// The usage is as follows:
// Ctl(CmdBuffer, bufferName)
// Ctl(CmdOpen, bufferName)
// Ctl(cmdClose, bufferName)
// Ctl(cmdLink, toBuffer, fromBuffer)
func (c *Client) Ctl(cmd CmdType, args ...string) (int, error) {
	nfid := c.clnt.FidAlloc()
	_, err := c.clnt.Walk(c.root, nfid, []string{"ctl"})
	if err != nil {
		return 0, err
	}

	c.clnt.Open(nfid, p.OAPPEND)
	defer c.clnt.Clunk(nfid)

	var data string
	switch cmd {
	case CmdBuffer:
		if len(args) > 1 {
			return 0, ErrBadArgs
		}

		data = fmt.Sprintf("buffer %s\x00", args[0])
	case CmdOpen:
		if len(args) > 1 {
			return 0, ErrBadArgs
		}

		data = fmt.Sprintf("open %s\x00", args[0])
	case CmdClose:
		if len(args) > 1 {
			return 0, ErrBadArgs
		}

		data = fmt.Sprintf("close %s\x00", args[0])
	case CmdLink:
		if len(args) != 2 {
			return 0, ErrBadArgs
		}

		data = fmt.Sprintf("link %s %s\x00", args[0], args[1])
	}
	return c.clnt.Write(nfid, []byte(data), 0)
}

// Tabs returns the contents of the `tabs` file for the server
func (c *Client) Tabs() ([]byte, error) {
	return getNamedFile(c, "tabs")
}

// Title returns the contents of the `title` file for a given buffer
func (c *Client) Title() ([]byte, error) {
	return getNamedFile(c, "title")
}

// Status returns the contents of the `status` file for a given buffer
func (c *Client) Status() ([]byte, error) {
	return getNamedFile(c, "status")
}

// Aside returns the contents of the `aside` file for a given buffer
func (c *Client) Aside() ([]byte, error) {
	return getNamedFile(c, "aside")
}

// Input appends the given data string to input
func (c *Client) Input(data []byte) (int, error) {
	nfid := c.clnt.FidAlloc()
	_, err := c.clnt.Walk(c.root, nfid, []string{"input"})
	if err != nil {
		return 0, err
	}

	c.clnt.Open(nfid, p.OAPPEND)
	defer c.clnt.Clunk(nfid)

	return c.clnt.Write(nfid, data, 0)
}

// Notifications returns and clears any pending notifications
func (c *Client) Notifications() ([]byte, error) {
	nfid := c.clnt.FidAlloc()
	_, err := c.clnt.Walk(c.root, nfid, []string{"notification"})
	if err != nil {
		return nil, err
	}

	c.clnt.Open(nfid, p.OREAD)
	defer c.clnt.Remove(nfid)
	//defer c.clnt.Clunk(nfid)

	return c.clnt.Read(nfid, 0, p.MSIZE)
}

// Feed returns a ReadCloser connected to `feed`. It's expected all reads
// will be read into a buffer with a size of MSIZE
// It is also expected for Feed to be called in its own thread
func (c *Client) Feed() (io.ReadCloser, error) {
	nfid := c.clnt.FidAlloc()

	_, err := c.clnt.Walk(c.root, nfid, []string{"feed"})
	if err != nil {
		return nil, err
	}

	c.clnt.Open(nfid, p.OREAD)

	data := make(chan []byte)
	done := make(chan struct{})

	go func() {
		var off uint64
		defer c.clnt.Clunk(nfid)

		for {

			b, err := c.clnt.Read(nfid, off, p.MSIZE)
			if err != nil {
				return
			}

			if len(b) > 0 {
				data <- b
				off += uint64(len(b))
			}

			select {
			case <-done:
				return
			default:
				continue
			}
		}

	}()

	f := &feed{
		data: data,
		done: done,
	}

	return f, nil

}

func getNamedFile(c *Client, name string) ([]byte, error) {
	nfid := c.clnt.FidAlloc()
	_, err := c.clnt.Walk(c.root, nfid, []string{name})
	if err != nil {
		return nil, err
	}

	c.clnt.Open(nfid, p.OREAD)
	defer c.clnt.Clunk(nfid)

	return c.clnt.Read(nfid, 0, p.MSIZE)
}
