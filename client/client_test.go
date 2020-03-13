package client

import (
	"io"
	"testing"
	"time"
)

func TestFeed(t *testing.T) {
	mc := NewMockClient("none")
	mc.Connect(1)
	mc.Attach()
	mc.Auth()

	f, err := mc.Feed()
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		for {
			b := make([]byte, MSIZE)

			_, err := f.Read(b)
			if err != nil && err != io.EOF {
				t.Error(err)
				return
			}
		}
	}()

	time.Sleep(time.Second * 10)
	f.Close()
}

func TestCommands(t *testing.T) {
	mc := NewMockClient("none")
	mc.Connect(1)
	mc.Attach()
	mc.Auth()
	mc.Input([]byte("Some text"))
	mc.Ctl(CmdOpen, "chicken")
	mc.Ctl(CmdClose, "chicken")
}
