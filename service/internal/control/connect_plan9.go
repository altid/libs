package control

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/altid/libs/service/commander"
)

var fd *os.File

func ConnectService(ctx context.Context, name string) (*Control, error) {
	b := make([]byte, 4)
	cfd, err := os.Open("/mnt/alt/clone")
	if err != nil {
		return nil, err
	}
	
	n, err := cfd.Read(b)
	if err != nil {
		return nil, err
	}

	// Instead, open up write only
	cf := path.Join("/mnt/alt", string(b[:n-1]), "ctl")
	wfd, err := os.OpenFile(cf, os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	rfd, err := os.Open(cf)
	if err != nil {
		return nil, err
	}

	fd = rfd

	ctl := &Control{
		cmds: make(chan *commander.Command),
		done: make(chan bool),
		errs: make(chan error),
		ctx: ctx,
		ctl: wfd,
	}

	// This creates /srv/$name, and returns our ctl file handle
	if _, err := fmt.Fprintf(wfd, "%s\n", name); err != nil {
		return nil, err
	}

	return ctl, nil
}

func (c *Control) ReadCommands() {
	// Read, don't stop on eof, just read again
	// from ctl.commander.FromString(data)
	// handle inputs as well, call handle, etc
	buf := make([]byte, 1024)
	for {
		n, err := fd.Read(buf)
		if err!= nil && err != io.EOF {
			c.errs <- err
			return
		}
		if n > 0 {
			// We really want to return this error in the future
			cmd, _ := c.commander.FromString(string(buf))
			c.cmds <- cmd
		}
	}
}
