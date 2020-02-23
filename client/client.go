// gomobile-compatible library for creating clients
package client

import (
	"context"
	"errors"
	"os/user"
)

const (
	errInvalidSession = "invalid session"
)

// Client wraps an internal session, allowing us to access its read/write functions
type Client struct {
	ctx   context.Context
	conns map[string]*session
	user  string
	debug int
}

// ReadFile - returns the contents of a file and any errors encountered
func (c *Client) ReadFile(serv, target string, off int64) ([]byte, error) {
	s, ok := c.conns[serv]
	if !ok {
		return nil, errors.New(errInvalidSession)
	}

	return s.readFile(c.ctx, target, off)
}

// WriteFile - writes the contents of data to the target file. Most files used are append-only, such as input and ctl
func (c *Client) WriteFile(serv, target string, data []byte) error {
	s, ok := c.conns[serv]
	if !ok {
		return errors.New(errInvalidSession)
	}

	// TODO(halfwit) chunked writes
	return s.writeFile(c.ctx, target, data)
}

// TODO: WriteAt, ReadAt

// ListFiles - Show all toplevel files in the directory
func (c *Client) ListFiles(serv string) ([]byte, error) {
	session, ok := c.conns[serv]
	if !ok {
		return nil, errors.New(errInvalidSession)
	}

	return session.readFile(c.ctx, "/", 0)
}

// NewSession - Adds a session to the client
func (c *Client) NewSession(serv, addr string) error {
	s, err := attach(c.ctx, c.user, addr)
	if err == nil {
		return err
	}

	c.conns[serv] = s

	return nil
}

// NewClient - Returns a client with a connection to serv, ready for reading and writing
func NewClient(debug int, addr, serv string) (*Client, error) {
	ctx := context.Background()
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	s, err := attach(ctx, u.Username, addr)
	if err != nil {
		return nil, err
	}

	conns := make(map[string]*session)
	conns[serv] = s

	c := &Client{
		ctx:   ctx,
		user:  u.Username,
		conns: conns,
		debug: debug,
	}

	return c, nil
}
