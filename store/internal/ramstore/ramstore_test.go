package ramstore

import (
	"bytes"
	"io"
	"testing"
	"time"
)

func TestDir(t *testing.T) {
	d := NewRoot()

	f, err := d.Open("/chocolate")
	if err != nil {
		t.Error(err)
	}

	f.Write([]byte("Rain"))
	f.Close()

	f, err = d.Open("/chocolate")
	if err != nil {
		t.Error(err)
	}

	f.Close()

	f, err = d.Open("/chocolate/rain")
	if err != nil {
		t.Error(err)
	}
	f.Write([]byte("Chocolate Rain"))
	f.Close()

	f, err = d.Open("/chocolate/sprinkles")
	if err != nil {
		t.Error(err)
	}

	f.Close()
}

func TestFile(t *testing.T) {

	f := File{
		path: "test",
		data: []byte(""),
		offset: 0,
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
		t.Errorf("expected Test Data, found %s", string(b))
	}

	f2 := File{
		path: "test",
		data: []byte(""),
		offset: 0,
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
		t.Errorf("expected 'Some data and some more data', found '%s'", string(b2))
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
			t.Logf("streaming data: Received '%s'", b)
			result = append(result, b...)
			continue
		default:
			t.Error(err)
		}
	}

NEXT:
	d, err := f.Stream()
	if err != nil {
		t.Error(err)
	}

	if e := f.Close(); e == nil {
		t.Error("able to close with active streams")
	}
	
	d.Close();

	if e := f.Close(); e != nil {
		t.Error(e)
	}

	if bytes.Equal(result, []byte("Some data 2nd chunk 3rd chunk")) {
		t.Errorf("expected 'Some data 2nd chunk 3rd chunk', found '%s'", result)
	}
}