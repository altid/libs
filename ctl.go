package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
)

func init() {
	s := &fileHandler{
		fn:   getCtl,
		stat: getCtlStat,
	}
	addFileHandler("/ctl", s)
}

type ctl struct {
	commands chan *cmd
	off      int64
	size     int64
	uuid     int64
	data     []byte
	path     string
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
	if err != nil {
		return 0, errors.New("nil or empty command received")
	}

	value, err := buff.ReadString('\n')
	if err != nil && err != io.EOF {
		return 0, err
	}

	value = value[:len(value)-1]

	switch command {
	case "refresh":
		c.commands <- &cmd{
			uuid:  c.uuid,
			key:   reloadCmd,
			value: value,
		}
	case "buffer ":
		c.commands <- &cmd{
			uuid:  c.uuid,
			key:   bufferCmd,
			value: value,
		}

		return len(p), nil
	case "close ":
		c.commands <- &cmd{
			uuid:  c.uuid,
			key:   closeCmd,
			value: value,
		}
	case "link ":
		c.commands <- &cmd{
			uuid:  c.uuid,
			key:   linkCmd,
			value: value,
		}
	case "open ":
		c.commands <- &cmd{
			uuid:  c.uuid,
			key:   openCmd,
			value: value,
		}
	}

	fp, err := os.OpenFile(c.path, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return 0, err
	}

	defer fp.Close()

	return fp.Write(p)
}

func (c *ctl) Close() error { return nil }
func (c *ctl) Uid() string  { return defaultUID }
func (c *ctl) Gid() string  { return defaultGID }

func getCtl(msg *message) (interface{}, error) {
	fp := path.Join(*inpath, msg.svc.name, "ctl")

	buff, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	c := &ctl{
		uuid:     msg.uuid,
		data:     buff,
		size:     int64(len(buff)),
		commands: msg.svc.commands,
		path:     fp,
	}

	return c, nil
}

func getCtlStat(msg *message) (os.FileInfo, error) {
	return os.Lstat(path.Join(*inpath, msg.svc.name, "ctl"))
}
