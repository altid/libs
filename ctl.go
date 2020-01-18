package main

import (
	"bytes"
	"errors"
	"io"
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
	state chan *update
	path  string
}

func (c *ctl) WriteAt(p []byte, off int64) (int, error) {
	buff := bytes.NewBuffer(p)
	command, err := buff.ReadString(' ')
	if err != nil {
		return 0, errors.New("Nil or empty command received")
	}
	switch command {
	case "buffer ":
		value, err := buff.ReadString('\n')
		if err != io.EOF {
			return 0, err
		}
		c.state <- &update{
			key:   bufferUpdate,
			value: value,
		}
		return 0, nil
	default:
		fp, err := os.OpenFile(c.path, os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return 0, err
		}
		defer fp.Close()
		return fp.Write(p)
	}
}

func (c *ctl) Close() error { return nil }

func getCtl(msg *message) (interface{}, error) {
	c := &ctl{
		state: msg.state,
		path:  path.Join(*inpath, msg.service, msg.buff, "ctl"),
	}
	if _, err := os.Stat(c.path); err != nil {
		return nil, err
	}
	return c, nil
}

// We should be able to get away with sending back a normal stat
func getCtlStat(msg *message) (os.FileInfo, error) {
	fp := path.Join(*inpath, msg.service, msg.buff, "ctl")
	return os.Stat(fp)
}
