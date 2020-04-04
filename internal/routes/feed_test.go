package routes

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestFeed(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}

	event := make(chan struct{}, 5)
	fh := NewFeed()
	fh.feeds[0] = event

	f := &feed{
		fh:   fh,
		uuid: 0,
		buff: "test",
		path: tmpfile.Name(),
	}

	// Set our initial data
	tmpfile.WriteString("Line of text\n")
	b := make([]byte, 145)
	n, _ := f.ReadAt(b, 0)

	go func(off int64) {
		for i := 0; i < 60; i++ {
			off += readLine(f, t, off)
		}
	}(int64(n))

	sendFeeds(tmpfile, fh)
	tmpfile.Close()
	os.Remove(tmpfile.Name())
}

// We send with a timeout, so we can read a line at a time
// We don't _need_ to read a line at a time at all in normal use
func sendFeeds(tmpfile *os.File, fh *FeedHandler) {
	for i := 0; i < 60; i++ {
		tmpfile.WriteString("Line of text\n")
		fh.Send(0)
		time.Sleep(time.Millisecond * 20)
	}

	// Let read catch up
	time.Sleep(time.Millisecond * 20)
	fh.Done(0)
}

func readLine(f *feed, t *testing.T, off int64) int64 {
	// Try to read more, so we don't get extra feed items
	b := make([]byte, 17)
	n, err := f.ReadAt(b, off)
	if err != nil {
		t.Error(err)
	}

	if string(b[:13]) != "Line of text\n" {
		t.Error("Strings did not match")
	}

	return int64(n)
}
