package command

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"testing"
)

func TestWriteOut(t *testing.T) {
	var data bytes.Buffer
	test := bufio.NewWriter(&data)

	a := New(0, LinkCmd, nil, "foo", "bar")
	if e := a.WriteOut(test); e != nil {
		t.Error(e)
	}

	test.Flush()

	if data.String() != `link "foo" "bar"`+"\n" {
		t.Error("Incorrect link output")
	}

	data.Reset()
	b := New(0, OpenCmd, nil, "foo", "bar")
	if e := b.WriteOut(test); e != nil {
		t.Error(e)
	}

	test.Flush()
	if data.String() != `open "foo" "bar"`+"\n" {
		t.Error("Incorrect open output")
	}

	c := New(0, CloseCmd, nil, "foo", "bar")
	if e := c.WriteOut(ioutil.Discard); e != nil {
		t.Error(e)
	}

	d := New(0, BufferCmd, nil, "foo", "bar")
	if e := d.WriteOut(ioutil.Discard); e != nil {
		t.Error(e)
	}

	data.Reset()
	f := New(0, QuitCmd, nil)
	if e := f.WriteOut(test); e != nil {
		t.Error(e)
	}

	test.Flush()
	if data.String() != `quit`+"\n" {
		t.Error("quit command improperly formatted")
	}

	g := New(0, ReloadCmd, nil)
	if e := g.WriteOut(ioutil.Discard); e == nil {
		t.Error("reload should return error on WriteOut")
	}

	h := New(0, OtherCmd, []byte("some data"))
	if e := h.WriteOut(ioutil.Discard); e != nil {
		t.Error(e)
	}
}
