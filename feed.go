package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"path"
)

// feed files are special in that they're blocking
type feed struct {
	buff bytes.Buffer
	done chan struct{}
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

	n = copy(b, f.buff.Next(len(b)))
	return
}

func (f *feed) Close() error {
	close(f.done)
	return nil
}

func getLines(msg *message, buff bytes.Buffer, feed, done chan struct{}) {
	fp, err := os.Open(path.Join(*inpath, msg.svc.name, msg.buff))
	if err != nil {
		log.Println(err)
		return
	}

	defer fp.Close()

	for {
		var b []byte
		select {
		case <-feed:
			n, err := fp.Read(b)
			if err != nil && err != io.EOF {
				buff.Reset()
				break
			}
			buff.Write(b[:n])
		case <-done:
			buff.Reset()
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

	go getLines(msg, buff, msg.svc.feed, done)

	return f, nil
}

func getFeedStat(msg *message) (os.FileInfo, error) {
	return os.Stat(path.Join(*inpath, msg.svc.name, msg.buff, "feed"))
}
