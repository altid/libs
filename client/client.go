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
	fids   map[string]*clnt.Fid
	addr   string
	buffer string
	clnt   *clnt.Clnt
}

// NewClient returns an authenticated client
func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
		fids: make(map[string]*clnt.Fid),
	}
}

// Connect performs the network dial for the connection
func (c *Client) Connect() (err error) {
	dial := fmt.Sprintf("%s:564", c.addr)

	conn, err := net.Dial("tcp", dial)
	if err != nil {
		return err
	}

	c.clnt, err = clnt.Connect(conn, p.MSIZE, false)
	if err != nil {
		return err
	}

	return
}

// Attach is called after optionally calling Auth
func (c *Client) Attach() (err error) {
	afid, ok := c.fids["auth"]
	if !ok {
		afid = nil
	}

	root, err := c.clnt.Attach(afid, user.Current(), "/")
	if err != nil {
		return err
	}

	c.fids["root"] = root

	return nil
}

// Auth is optionally called after Connect to authenticate with the server
func (c *Client) Auth() error {
	afid, err := c.clnt.Auth(user.Current(), "/")
	if err != nil {
		return err
	}

	c.fids["auth"] = afid
	return nil
}

// Tabs returns the contents of a servers' tabs file
func (c *Client) Tabs() ([]byte, error) {
	nfid, err := c.getFid("tabs", p.OREAD)
	if err != nil {
		return nil, err
	}

	return c.clnt.Read(nfid, 0, p.MSIZE)
}

// Buffer attempts to switch to a named buffer
func (c *Client) Buffer(name string) (n int, err error) {
	nfid, err := c.getFid("ctl", p.OAPPEND)
	if err != nil {
		return 0, err
	}

	data := fmt.Sprintf("buffer %s\x00", name)
	return c.clnt.Write(nfid, []byte(data), 0)
}

// Feed returns a ReadCloser connected to `feed`. It's expected all reads
// will be read into a buffer with a size of MSIZE
// It is also expected for Feed to be called in its own thread
func (c *Client) Feed() (io.ReadCloser, error) {
	nfid, err := c.getFid("/feed", p.OREAD)
	if err != nil {
		return nil, err
	}

	data := make(chan []byte)
	done := make(chan struct{})

	go func() {
		var off uint64

		for {

			b, err := c.clnt.Read(nfid, off, p.MSIZE)
			if err != nil && err != io.EOF {
				return
			}

			if len(b) > 0 {
				data <- b
				off += uint64(len(b))
			}

			select {
			case <-done:
				break
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

func (c *Client) getFid(name string, mode uint8) (*clnt.Fid, error) {
	f, ok := c.fids[name]
	if !ok {
		nfid, err := c.clnt.FWalk(name)
		if err != nil {
			return nil, err
		}

		c.fids[name] = nfid
		c.clnt.Open(nfid, mode)

		return nfid, nil
	}

	return f, nil
}
