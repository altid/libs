package store

import (
	"io"
	"testing"
)

func TestRamstore(t *testing.T) {
	rs := NewRamstore(true)
	if e := rs.Mkdir("/chicken"); e != nil {
		t.Error(e)
	}
	_, err := rs.Open("/chicken/cluck")
	if err != nil {
		t.Error(err)
	}

	f, err := rs.Open("/check")
	if err != nil {
		t.Error(err)
	}
	if _, e := f.Write([]byte("Hello")); e != nil {
		t.Error(e)
	}
	if _, e := f.Seek(0, io.SeekStart); e != nil {
		t.Error(e)
	}
	b := make([]byte, 5)
	if _, e := f.Read(b); e != nil && e != io.EOF {
		t.Error(e)
	}
	if string(b) != "Hello" {
		t.Error("Strings do not match")
	}
	//f.Close()
}
