package command

import (
	"io/ioutil"
	"testing"
)

func TestFromBytes(t *testing.T) {
	if c, e := FromBytes(0, "mybuffer", []byte("open foo")); e != nil || c.Args[0] != "foo" {
		t.Errorf("from bytes failed on open foo %v", e)
	}

	if c, e := FromBytes(0, "mybuffer", []byte("close bar")); e != nil || c.Args[0] != "bar" {
		t.Errorf("from bytes failed on close bar %v", e)
	}

	if c, e := FromBytes(0, "mybuffer", []byte("buffer foo")); e != nil || c.Args[0] != "foo" {
		t.Errorf("from bytes failed on buffer foo %v", e)
	}

	if c, e := FromBytes(0, "mybuffer", []byte("reload")); e != nil || c.From != "mybuffer" {
		t.Errorf("from bytes failed on reload %v", e)
	}

	if c, e := FromBytes(0, "mybuffer", []byte("quit")); e != nil || c.CmdType != QuitCmd {
		t.Errorf("from bytes failed on quit %v", e)
	}

	if c, e := FromBytes(0, "mybuffer", []byte("link foo bar")); e != nil || len(c.Args) != 2 || c.Args[0] != "foo" || c.Args[1] != "bar" {
		t.Errorf("from bytes failed on link %v", e)
	}

	if c, e := FromBytes(0, "mybuffer", []byte("some other command")); e != nil || c.CmdType != OtherCmd {
		t.Errorf("from bytes failed on other test %v", e)
	}

func TestWriteOut(t *testing.T) {
	a := New(0, LinkCmd, "foo", "bar")
	if e := a.WriteOut(ioutil.Discard); e != nil {
		t.Error(e)
	}

	b := New(0, OpenCmd, "bar", "foo")
	if e := b.WriteOut(ioutil.Discard); e != nil {
		t.Error(e)
	}

	c := New(0, CloseCmd, "bar", "foo")
	if e := c.WriteOut(ioutil.Discard); e != nil {
		t.Error(e)
	}

	d := New(0, BufferCmd, "foo", "bar")
	if e := d.WriteOut(ioutil.Discard); e != nil {
		t.Error(e)
	}

	f := New(0, QuitCmd, "foo")
	if e := f.WriteOut(ioutil.Discard); e != nil {
		t.Error(e)
	}

	g := New(0, ReloadCmd, "foo")
	if e := g.WriteOut(ioutil.Discard); e == nil {
		t.Error("reload should return error on WriteOut")
	}

	h := New(0, OtherCmd, "foo", "cmd")
	if e := h.WriteOut(ioutil.Discard); e != nil {
		t.Error(e)
	}
}
