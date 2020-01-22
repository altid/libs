package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"
)

func init() {
	s := &fileHandler{
		fn:   getCtl,
		stat: getCtlStat,
	}
	addFileHandler("/ctl", s)
}

type ctl struct {
	state   chan *update
	modTime time.Time
	off     int64
	size    int64
	data    []byte
	path    string
}

func (c *ctl) ReadAt(b []byte, off int64) (n int, err error) {
	n = copy(b, c.data[off:])
	if int64(n)+off > c.size {
		return n, io.EOF
	}
	return
}

func (c *ctl) WriteAt(p []byte, off int64) (int, error) {
	c.modTime = time.Now().Truncate(time.Hour)
	c.off += off + int64(len(p))
	buff := bytes.NewBuffer(p)
	command, err := buff.ReadString(' ')
	if err != nil {
		return 0, errors.New("Nil or empty command received")
	}
	value, err := buff.ReadString('\n')
	if err != io.EOF {
		return 0, err
	}

	switch command {
	case "buffer ":
		c.state <- &update{
			key:   bufferUpdate,
			value: value,
		}
		return len(p), nil
	case "close ":
		c.state <- &update{
			key:   closeUpdate,
			value: value,
		}
	case "link ":
		c.state <- &update{
			key:   linkUpdate,
			value: value,
		}
	case "open ":
		c.state <- &update{
			key:   openUpdate,
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
	fp := path.Join(*inpath, msg.service, "ctl")
	buff, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}
	c := &ctl{
		data:    buff,
		size:    int64(len(buff)),
		modTime: time.Now(),
		state:   msg.state,
		path:    fp,
	}
	return c, nil
}

// We should be able to get away with sending back a normal stat
func getCtlStat(msg *message) (os.FileInfo, error) {
	fp := path.Join(*inpath, msg.service, "ctl")
	return os.Lstat(fp)
}
