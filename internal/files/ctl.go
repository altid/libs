package files

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/altid/server/command"
	"github.com/altid/server/files"
)

// CtlHandler with access to Command
type CtlHandler struct {
	uuid uint32
	cmds chan *command.Command
	// Put the command here
}

func NewCtl(uuid, cmds chan *command.Command) *CtlHandler {	return &CtlHandler{uuid, cmds}}

func (ch *CtlHandler) Normal(msg *files.Message) (interface{}, error) {
	fp := path.Join(msg.Service, msg.Buffer, "ctl")

	buff, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	c := &ctl{
		cmds: ch,
		uuid: msg.UUID,
		data: buff,
		size: int64(len(buff)),
		path: fp,
	}

	return c, nil
}

func (*CtlHandler) Stat(msg *files.Message) (os.FileInfo, error) {
	return os.Lstat(path.Join(msg.Service, msg.Buffer, "ctl"))
}

type ctl struct {
	cmds chan *command.Command
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

	value = value[:len(value)-1]

	switch cmd {
	case "refresh":
		c.cmds <- command.New(c.uuid, command.ReloadCmd)
	case "buffer":
		c.cmds <- command.New(c.uuid, command.BufferCmd, value)
	case "close":
		c.cmds <- command.New(c.uuid, command.CloseCmd, value)
	case "link":
		t := strings.Fields(value)
		if len(t) != 2 {
			return 0, errors.New("link requires two arguments")
		}
		c.cmds <- command.New(c.uuid, command.LinkCmd, t[0], t[1])
	case "open":
		c.cmds <- command.New(c.uuid, command.OpenCmd, value)

	default:
		c.cmds <- command.New(c.uuid, command.OtherCmd, fmt.Sprintf("%s %s", cmd, value))
	}

	return len(p), nil
}

func (c *ctl) Close() error { return nil }
