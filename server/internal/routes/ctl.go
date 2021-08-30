package routes

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/altid/server/command"
	"github.com/altid/server/internal/message"
)

// CtlHandler with access to Command
type CtlHandler struct {
	cmds chan *command.Command
}

func NewCtl(cmds chan *command.Command) *CtlHandler { return &CtlHandler{cmds} }

func (ch *CtlHandler) Normal(msg *message.Message) (interface{}, error) {
	fp := path.Join(msg.Service, "ctl")

	buff, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	c := &ctl{
		uuid: msg.UUID,
		cmds: ch.cmds,
		data: buff,
		size: int64(len(buff)),
		path: fp,
		curr: msg.Buffer,
	}

	return c, nil
}

func (*CtlHandler) Stat(msg *message.Message) (os.FileInfo, error) {
	return os.Lstat(path.Join(msg.Service, "ctl"))
}

type ctl struct {
	cmds chan *command.Command
	off  int64
	size int64
	uuid uint32
	data []byte
	path string
	curr string
}

func (c *ctl) ReadAt(b []byte, off int64) (n int, err error) {
	n = copy(b, c.data[off:])
	if int64(n)+off > c.size {
		return n, io.EOF
	}

	return
}

func (c *ctl) WriteAt(p []byte, off int64) (int, error) {
	buff := bytes.NewBuffer(p)

	cmd, err := buff.ReadString(' ')
	if err != nil && err != io.EOF {
		return 0, errors.New("nil or empty command received")
	}

	cmd = strings.TrimSuffix(cmd, " ")

	value, err := buff.ReadString('\n')
	if err != nil && err != io.EOF {
		return 0, err
	}

	value = strings.TrimSuffix(value, "\n")

	switch cmd {
	case "refresh":
		c.cmds <- command.New(c.uuid, command.ReloadCmd, c.curr, value)
	case "buffer":
		c.cmds <- command.New(c.uuid, command.BufferCmd, c.curr, value)
	case "close":
		c.cmds <- command.New(c.uuid, command.CloseCmd, c.curr, value)
	case "link":
		c.cmds <- command.New(c.uuid, command.LinkCmd, c.curr, value)
	case "open":
		c.cmds <- command.New(c.uuid, command.OpenCmd, c.curr, value)
	default:
		c.cmds <- command.New(c.uuid, command.OtherCmd, c.curr, cmd, value)
	}

	return len(p), nil
}

func (c *ctl) Close() error { return nil }
