package routes

import (
	"testing"

	"github.com/altid/server/command"
)

// Read and write to ctl

func TestCtl(t *testing.T) {
	cmds := make(chan *command.Command)
	ctl := &ctl{
		uuid: 0,
		cmds: cmds,
		data: []byte("test\n"),
		size: 5,
		curr: "test",
	}

	go sendCmds(t, ctl)

	resp := <-cmds
	if resp.CmdType != command.OpenCmd || resp.Args[0] != "foo" || resp.From != "test" {
		t.Error("Unable to parse open command")
	}

	resp = <-cmds
	if resp.CmdType != command.BufferCmd || resp.Args[0] != "bar" {
		t.Error("Unable to parse buffer command")
	}

	resp = <-cmds
	if resp.CmdType != command.OtherCmd || resp.Args[1] != "to do" {
		t.Error("Unable to parse arbitrary command")
	}
}

func sendCmds(t *testing.T, ctl *ctl) {
	if _, e := ctl.WriteAt([]byte("open foo\n"), 0); e != nil {
		t.Error(e)
	}

	if _, e := ctl.WriteAt([]byte("buffer bar"), 0); e != nil {
		t.Error(e)
	}

	if _, e := ctl.WriteAt([]byte("nothing to do"), 0); e != nil {
		t.Error(e)
	}
}
