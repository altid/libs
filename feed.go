package main

import (
	"errors"
	"io"
	"os"
	"path"
)

// feed files are special in that they're blocking
type feed struct {
	tailing  bool
	path     string
	incoming chan struct{}
	done     chan struct{}
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

	// Bit of a trick here to get a single event on a potentially closed pipe
	for range f.incoming {
		n, err = fp.ReadAt(p, off)
		if err == io.EOF {
			return n, nil
		}

		return
	}

	return 0, io.EOF
}

func (f *feed) WriteAt(p []byte, off int64) (int, error) {
	return 0, errors.New("writing to feed files is currently unsupported")
}

func (f *feed) Close() error { return nil }

func getFeed(msg *message) (interface{}, error) {
	done := make(chan struct{})
	f := &feed{
		path: path.Join(*inpath, msg.svc.name, msg.buff, "feed"),
		done: done,
	}

	for _, cl := range msg.svc.clients {
		if cl.uuid != msg.uuid {
			continue
		}

		f.incoming = cl.feed

		return f, nil
	}

	return nil, errors.New("unable to open file")
}

func getFeedStat(msg *message) (os.FileInfo, error) {
	return os.Lstat(path.Join(*inpath, msg.svc.name, msg.buff, "feed"))
}
