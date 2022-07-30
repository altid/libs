package ramstore

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"
)

func TestStore(t *testing.T) {

	f := File{
		path: "test",
		data: []byte(""),
		offset: 0,
		closed: false,
		streams: make(map[string]*Stream),
	}

	if _, e := f.Seek(0, io.SeekStart); e != nil {
		t.Error(e)
	}
	b := []byte("Test data")
	if _, e := f.Write(b); e != nil {
		t.Error(e)
	}

	if _, e := f.Seek(0, io.SeekStart); e != nil {
		t.Error(e)
	}

	if _, e := f.Read(b); e != nil && e != io.EOF {
		t.Error(e)
	}

	if err := f.Close(); err != nil {
		t.Error(err)
	}

	if string(b) != "Test data" {
		t.Errorf("Expected Test Data, found %s", string(b))
	}

	if _, e := f.Write([]byte(" more data")); e == nil {
		t.Error(fmt.Errorf("Should not write on a closed file"))
	}

	f2 := File{
		path: "test",
		data: []byte(""),
		offset: 0,
		closed: false,
		streams: make(map[string]*Stream),
	}

	if _, e := f2.Write([]byte("Some data ")); e != nil {
		t.Error(e)
	}

	if _, e := f2.Seek(0, io.SeekEnd); e != nil {
		t.Error(e)
	}

	if _, e := f2.Write([]byte("and some more data")); e != nil {
		t.Error(e)
	}

	b2 := make([]byte, 28)
	if _, e := f2.Read(b2); e != nil && e != io.EOF {
		t.Error(e)
	}

	if string(b2) != "Some data and some more data" {
		t.Errorf("Expected 'Some data and some more data', found '%s'", string(b2))
	}

	if e := f2.Close(); e != nil {
		t.Error(e)
	}
}

func TestStream(t *testing.T) {
	f := File{
		path: "test",
		data: []byte(""),
		offset: 0,
		closed: false,
		streams: make(map[string]*Stream),
	}

	f.Write([]byte("Some data "))

	c, err := f.Stream()
	if err != nil {
		t.Error(err)
	}

	b := make([]byte, 1024)

	go func(f File, c io.ReadCloser) {
		time.Sleep(time.Second)
		f.Write([]byte("2nd chunk "))
		time.Sleep(time.Second)
		f.Write([]byte("3rd chunk"))
		time.Sleep(time.Second)
		c.Close()
	}(f, c)

	result := []byte("")

	for {
		_, err := c.Read(b)
		switch(err) {
		case io.EOF:
			goto NEXT
		case nil:
			t.Logf("Streaming data: Received '%s'", b)
			result = append(result, b...)
			continue
		default:
			t.Error(err)
		}
	}

NEXT:
	if e := f.Close(); e != nil {
		t.Error(e)
	}

	if bytes.Equal(result, []byte("Some data 2nd chunk 3rd chunk")) {
		t.Errorf("Expected 'Some data 2nd chunk 3rd chunk', found '%s'", result)
	}
}