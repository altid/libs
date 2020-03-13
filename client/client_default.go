package client

import (
	"fmt"
	"io"
	"net"

	"github.com/knieriem/g/go9p/user"
	"github.com/lionkov/go9p/p"
	"github.com/lionkov/go9p/p/clnt"
)

type client struct {
	afid   *clnt.Fid
	root   *clnt.Fid
	addr   string
	buffer string
	clnt   *clnt.Clnt
}

func (c *client) connect(debug int) (err error) {
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

func (c *client) cleanup() {
	c.clnt.Unmount()
}

func (c *client) attach() error {
	root, err := c.clnt.Attach(c.afid, user.Current(), "/")
	if err != nil {
		return err
	}

	c.root = root

	return nil
}

func (c *client) auth() error {
	// TODO(halfwit): We want to flag in factotum use and hand it the afid
	afid, err := c.clnt.Auth(user.Current(), "/")
	if err != nil {
		return err
	}

	c.afid = afid
	return nil
}

func (c *client) ctl(cmd CmdType, args ...string) (int, error) {
	nfid := c.clnt.FidAlloc()
	_, err := c.clnt.Walk(c.root, nfid, []string{"ctl"})
	if err != nil {
		return 0, err
	}
	c.clnt.Open(nfid, p.OAPPEND)
	defer c.clnt.Clunk(nfid)

	data, err := runClientCtl(cmd, args...)
	if err != nil {
		return 0, err
	}

	return c.clnt.Write(nfid, data, 0)
}

func (c *client) tabs() ([]byte, error) {
	return getNamedFile(c, "tabs")
}

func (c *client) title() ([]byte, error) {
	return getNamedFile(c, "title")
}

func (c *client) status() ([]byte, error) {
	return getNamedFile(c, "status")
}

func (c *client) aside() ([]byte, error) {
	return getNamedFile(c, "aside")
}

func (c *client) input(data []byte) (int, error) {
	nfid := c.clnt.FidAlloc()
	_, err := c.clnt.Walk(c.root, nfid, []string{"input"})
	if err != nil {
		return 0, err
	}

	c.clnt.Open(nfid, p.OAPPEND)
	defer c.clnt.Clunk(nfid)

	return c.clnt.Write(nfid, data, 0)
}

func (c *client) notifications() ([]byte, error) {
	nfid := c.clnt.FidAlloc()
	_, err := c.clnt.Walk(c.root, nfid, []string{"notification"})
	if err != nil {
		return nil, err
	}

	c.clnt.Open(nfid, p.OREAD)
	defer c.clnt.Remove(nfid)

	return c.clnt.Read(nfid, 0, p.MSIZE)
}

func (c *client) feed() (io.ReadCloser, error) {
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

func getNamedFile(c *client, name string) ([]byte, error) {
	nfid := c.clnt.FidAlloc()
	_, err := c.clnt.Walk(c.root, nfid, []string{name})
	if err != nil {
		return nil, err
	}

	c.clnt.Open(nfid, p.OREAD)
	defer c.clnt.Clunk(nfid)

	return c.clnt.Read(nfid, 0, p.MSIZE)
}
