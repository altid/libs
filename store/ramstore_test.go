package store

import (
	"io"
	"reflect"
	"testing"
)

func TestRamstore(t *testing.T) {
	rs := NewRamstore(true)
	if e := rs.Mkdir("chicken"); e != nil {
		t.Error(e)
	}
	f, err := rs.Open("chicken/cluck")
	if err != nil {
		t.Error(err)
	}
	f.Write([]byte("Test data"))
	f.Close()
	f, err = rs.Open("chicken/cluck")
	if err != nil {
		t.Error(err)
	}
	f.Seek(0, io.SeekEnd)
	if _, e := f.Write([]byte(" Hello")); e != nil {
		t.Error(e)
	}
	if _, e := f.Seek(0, io.SeekStart); e != nil {
		t.Error(e)
	}
	b := make([]byte, 15)
	if _, e := f.Read(b); e != nil && e != io.EOF {
		t.Error(e)
	}
	if string(b) != "Test data Hello" {
		t.Error("Strings do not match")
	}
	if f.Name() != "chicken/cluck" {
		t.Error("Expected file name different from actual")
	}
	f.Close()
}

func TestRamstoreList(t *testing.T) {
	// Array to test against
	n := []string{"ptarmigan/leg", "swan"}

	rs := NewRamstore(true)
	if e := rs.Mkdir("chicken"); e != nil {
		t.Error(e)
	}
	if e := rs.Mkdir("duck"); e != nil {
		t.Error(e)
	}
	if e := rs.Mkdir("ptarmigan"); e != nil {
		t.Error(e)
	}
	if _, e := rs.Open("ptarmigan/leg"); e != nil {
		t.Error(e)
	}
	if _, e := rs.Open("swan"); e != nil {
		t.Error(e)
	}
	l := rs.List()
	if ! reflect.DeepEqual(l, n) {
		t.Error("string arrays do not match")
	}
	t.Log(l, n)
}

