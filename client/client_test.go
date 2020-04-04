package client

import (
	"testing"
)

func TestManager(t *testing.T) {
	h := Manager{}

	client1 := h.Client(0)

	if client1 == nil {
		t.Error("client was nil")
		return
	}

	client1.SetBuffer("banana")

	h.Client(0)

	client2 := h.Client(0)
	if client2 == nil {
		t.Error("client2 was nil")
		return
	}

	client2.SetBuffer("orange")

	if len(h.List()) != 3 {
		t.Error("unable to create client")
	}

	h.Remove(client2.UUID)
	client1 = h.Client(client1.UUID)
	if client1 == nil {
		t.Error("unable to retrieve client")
		return
	}

	if !client1.Active {
		t.Error("unable to retrieve client")
	}

	client2 = h.Client(0)
	if client2 == nil || client2.Active {
		t.Error("remove failed, client maintained state across re-creation")
	}

	h.Remove(client2.UUID)

	if len(h.List()) != 2 {
		t.Error("unable to remove item from list")
	}

	client1.SetBuffer("foo")
	client1.SetBuffer("bar")

	if len(client1.History()) != 2 || client1.History()[0] != "banana" {
		t.Error("incorrect history")
	}

	if e := client1.Previous(); e != nil {
		t.Error(e)
	}

	if client1.Current() != "foo" {
		t.Error("Unable to return to previous buffer")
	}
}
