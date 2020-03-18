package files

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/altid/server/files"
)

type ctlHandler struct {
	// Put the command here
}

func (*ctlHandler) Normal(msg *files.Message) (interface{}, error) {
	fp := path.Join(msg.Service, msg.Buffer, "ctl")

	buff, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	c := &ctl{
		uuid: msg.UUID,
		data: buff,
		size: int64(len(buff)),
		path: fp,
	}

	return c, nil
}

func (*ctlHandler) Stat(msg *files.Message) (os.FileInfo, error) {
	return os.Lstat(path.Join(msg.Service, msg.Buffer, "ctl"))
}

type ctl struct {
	off  int64
	size int64
	uuid int64
	data []byte
	path string
}

func (c *ctl) ReadAt(b []byte, off int64) (n int, err error) {
	n = copy(b, c.data[off:])
	if int64(n)+off > c.size {

		return n, io.EOF
	}

	return
}

func (c *ctl) WriteAt(p []byte, off int64) (int, error) {
	c.off += off + int64(len(p))
	buff := bytes.NewBuffer(p)

	command, err := buff.ReadString(' ')
	if err != nil && err != io.EOF {
		return 0, errors.New("nil or empty command received")
	}

	command = strings.TrimSuffix(command, " ")

	value, err := buff.ReadString('\n')
	if err != nil && err != io.EOF {
		return 0, err
	}

	value = value[:len(value)-1]

	c.data = append(c.data, p...)

	// TODO: We want to send the cmd on the internal cmd handler for ctl
	switch command {
	case "refresh":
		//c.cmds.Send(uuid, command.Reload, value)
	case "buffer":
		//c.cmds.Send(uuid, command.Buffer, value)
		return len(p), nil
	case "close":
		//c.cmds.Send(uuid, command.Close, value)
	case "link":
		//c.cmds.Send(uuid, command.Link, value)
	case "open":
		//c.cmds.Send(uuid, command.Open, value)
		if value == "none" {
			return len(p), nil
		}
	}

	return 0, errors.New("No such command")
}

func (c *ctl) Close() error { return nil }
