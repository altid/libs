package client

import (
	"fmt"
	"io"
	"net"

	"github.com/knieriem/g/go9p/user"
	"github.com/lionkov/go9p/p"
	"github.com/lionkov/go9p/p/clnt"
)

// MSIZE - maximum size for a message
const MSIZE = p.MSIZE

// Client represents a 9p client session
type Client struct {
	addr   string
	buffer string
	clnt   *clnt.Clnt
	fid    *clnt.Fid
	afid   *clnt.Fid
}

// NewClient returns an authenticated client
func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

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

func (c *Client) Attach() (err error) {
	c.fid, err = c.clnt.Attach(c.afid, user.Current(), "/")
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Auth() (err error) {
	c.afid, err = c.clnt.Auth(user.Current(), "/")
	if err != nil {
		return err
	}

	return nil
}

// May not be usable for mobile, but still usable for normal clients
func (c *Client) Open(path string) (io.ReadWriteCloser, error) {
	return c.clnt.FOpen(path, p.ORDWR)
}

func (c *Client) Tabs() ([]byte, error) {
	nfid := c.clnt.FidAlloc()
	if _, e := c.clnt.Walk(c.fid, nfid, []string{"tabs"}); e != nil {
		return nil, e
	}

	c.fid = nfid
	if e := c.clnt.Open(nfid, p.OREAD); e != nil {
		return nil, e
	}

	return c.clnt.Read(nfid, 0, p.MSIZE)
}

// Buffer attempts to switch to a named buffer
func (c *Client) Buffer(name string) (n int, err error) {
	nfid, err := c.clnt.FWalk("/ctl")
	if err != nil {
		return 0, err
	}

	if e := c.clnt.Open(nfid, p.OAPPEND); e != nil {
		return 0, e
	}

	data := fmt.Sprintf("buffer %s\x00", name)
	return c.clnt.Write(nfid, []byte(data), 0)
}

// Feed returns a ReadCloser connected to `feed`. It's expected all reads
// will be read into a buffer with a size of MSIZE
// It is also expected for Feed to be called in its own thread
func (c *Client) Feed() (io.ReadCloser, error) {
	fd, err := c.clnt.FOpen("/feed", p.OREAD)
	if err != nil {
		return nil, err
	}

	data := make(chan []byte)
	done := make(chan struct{})

	go func() {
		var off int64

		for {
			b := make([]byte, p.MSIZE)

			n, err := fd.ReadAt(b, off)
			if err != nil && err != io.EOF {
				return
			}

			if n > 0 {
				data <- b[:n]
				off += int64(n)
			}
		}

	}()

	f := &feed{
		data: data,
		done: done,
	}

	return f, nil

}

type feed struct {
	data chan []byte
	done chan struct{}
}

func (f *feed) Read(b []byte) (n int, err error) {
	select {
	case in := <-f.data:
		n = copy(b, in)
		return
	case <-f.done:
		return 0, io.EOF
	}
}

func (f *feed) Close() error {
	f.done <- struct{}{}
	return nil
}
