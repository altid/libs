package main

import (
	"errors"
	"io"
	"os"
	"path"
	"time"
)

// feed files are special in that they're blocking
type feed struct {
	tailing  bool
	path     string
	buff     string
	incoming chan struct{}
	done     chan struct{}
	debug    func(string, ...interface{})
}

func init() {
	s := &fileHandler{
		fn:   getFeed,
		stat: getFeedStat,
	}
	addFileHandler("/feed", s)
}

func (f *feed) ReadAt(p []byte, off int64) (n int, err error) {
	fp, err := os.Open(f.path)
	if err != nil {
		return 0, err
	}

	defer fp.Close()

	if !f.tailing {
		n, err = fp.ReadAt(p, off)
		if err != nil && err != io.EOF {
			return
		}

		if err == io.EOF {
			f.tailing = true
		}
		return n, nil
	}

	for range f.incoming {
		f.debug("feed event on %s", f.buff)
		n, err = fp.ReadAt(p, off)
		if err == io.EOF {
			time.Sleep(time.Millisecond * 150)
			return n, nil
		}

		return
	}

	f.debug("feed closed %s", f.buff)
	return 0, io.EOF
}

func (f *feed) WriteAt(p []byte, off int64) (int, error) {
	f.debug("Attempted write on feed")
	return 0, errors.New("writing to feed files is currently unsupported")
}

func (f *feed) Close() error { return nil }

func getFeed(msg *message) (interface{}, error) {
	done := make(chan struct{})
	f := &feed{
		path:  path.Join(*inpath, msg.svc.name, msg.buff, "feed"),
		buff:  path.Join(msg.svc.name, msg.buff),
		done:  done,
		debug: msg.svc.debug,
	}

	cl := msg.svc.clients[msg.uuid]
	f.incoming = cl.feed

	f.debug("feed started %s", f.buff)
	return f, nil

}

func getFeedStat(msg *message) (os.FileInfo, error) {
	return os.Lstat(path.Join(*inpath, msg.svc.name, msg.buff, "feed"))
}
