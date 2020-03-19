package client

import (
	"testing"
)

func TestManager(t *testing.T) {
	h := Manager{}

	client1 := h.Client(0)

	// Should also set Active
	client1.SetBuffer("banana")

	h.Client(0)
	client2 := h.Client(0)
	client2.SetBuffer("orange")

	if len(h.List()) != 3 {
		t.Error("unable to create client")
	}

	h.Remove(client2.UUID)
	client1 = h.Client(client1.UUID)

	if !client1.Active {
		t.Error("unable to retrieve client")
	}

	client2 = h.Client(client2.UUID)
	if client2.Active {
		t.Error("remove failed, client maintained state across re-creation")
	}

	h.Remove(client2.UUID)

	if len(h.List()) != 2 {
		t.Error("unable to remove item from list")
	}
}
