package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"path"
	"time"
)

// feed files are special in that they're blocking
type feed struct {
	incoming chan struct{}
	buff     bytes.Buffer
	done     chan struct{}
}

func init() {
	s := &fileHandler{
		fn:   getFeed,
		stat: getFeedStat,
	}
	addFileHandler("/feed", s)
}

func (f *feed) ReadAt(b []byte, off int64) (n int, err error) {
	if f.buff.Len() == 0 {
		return 0, io.EOF
	}

	copy(b, f.buff.Next(1024))
	time.Sleep(200 * time.Millisecond)

	return
}

func (f *feed) Close() error {
	close(f.done)
	return nil
}

func (f *feed) getLines(msg *message) {
	fp, err := os.Open(path.Join(*inpath, msg.svc.name, msg.buff, "feed"))
	if err != nil {
		log.Println(err)
		return
	}

	defer fp.Close()

	for {
		b := make([]byte, 1024)

		n, err := fp.Read(b)
		if err != nil || n == 0 {
			break
		}

		f.buff.Write(b)
	}

	for {
		b := make([]byte, 1024)

		select {
		case <-f.incoming:
			n, err := fp.Read(b)
			if err == io.EOF || n == 0 {
				continue
			}
			if err != nil {
				f.buff.Reset()
				break
			}

			f.buff.Write(b)
		case <-f.done:
			f.buff.Reset()
			break
		}
	}
}

func getFeed(msg *message) (interface{}, error) {
	var buff bytes.Buffer

	done := make(chan struct{})
	f := &feed{
		buff: buff,
		done: done,
	}

	for _, cl := range msg.svc.clients {
		if cl.uuid != msg.uuid {
			continue
		}

		f.incoming = cl.feed
		go f.getLines(msg)

		return f, nil
	}

	return nil, errors.New("unable to open file")
}

func getFeedStat(msg *message) (os.FileInfo, error) {
	return os.Lstat(path.Join(*inpath, msg.svc.name, msg.buff, "feed"))
}
