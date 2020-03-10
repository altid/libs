package client

import (
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

// Tabs returns the contents of a servers' tabs file
func (c *Client) Tabs() ([]byte, error) {
	nfid := c.clnt.FidAlloc()
	_, err := c.clnt.Walk(c.root, nfid, []string{"tabs"})
	if err != nil {
		return nil, err
	}

	c.clnt.Open(nfid, p.OREAD)
	defer c.clnt.Clunk(nfid)

	return c.clnt.Read(nfid, 0, p.MSIZE)
}

// Buffer attempts to switch to a named buffer
func (c *Client) Buffer(name string) (int, error) {
	nfid := c.clnt.FidAlloc()
	_, err := c.clnt.Walk(c.root, nfid, []string{"ctl"})
	if err != nil {
		return 0, err
	}

	c.clnt.Open(nfid, p.OAPPEND)
	defer c.clnt.Clunk(nfid)

	data := fmt.Sprintf("buffer %s\x00", name)
	return c.clnt.Write(nfid, []byte(data), 0)
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

