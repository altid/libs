package command

import (
<<<<<<< HEAD:command/command_test.go
	"bufio"
	"bytes"
=======
>>>>>>> dev:internal/command/command_test.go
	"io/ioutil"
	"testing"
)

<<<<<<< HEAD:command/command_test.go
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
=======
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
>>>>>>> dev:internal/command/command_test.go
	}

<<<<<<< HEAD:command/command_test.go
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
=======
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
>>>>>>> dev:internal/command/command_test.go
	if e := c.WriteOut(ioutil.Discard); e != nil {
		t.Error(e)
	}

<<<<<<< HEAD:command/command_test.go
	d := New(0, BufferCmd, nil, "foo", "bar")
=======
	d := New(0, BufferCmd, "foo", "bar")
>>>>>>> dev:internal/command/command_test.go
	if e := d.WriteOut(ioutil.Discard); e != nil {
		t.Error(e)
	}

<<<<<<< HEAD:command/command_test.go
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
=======
	f := New(0, QuitCmd, "foo")
	if e := f.WriteOut(ioutil.Discard); e != nil {
		t.Error(e)
	}

	g := New(0, ReloadCmd, "foo")
>>>>>>> dev:internal/command/command_test.go
	if e := g.WriteOut(ioutil.Discard); e == nil {
		t.Error("reload should return error on WriteOut")
	}

	h := New(0, OtherCmd, "foo", "cmd")
	if e := h.WriteOut(ioutil.Discard); e != nil {
		t.Error(e)
	}
}
