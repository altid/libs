package client

import (
	"testing"
)

func TestManager(t *testing.T) {
	h := Manager{}

	client1 := h.Create(1234)

	// Should also set Active
	client1.SetBuffer("banana")

	h.Create(4321)
	client2 := h.Create(1111)
	client2.SetBuffer("orange")

	if len(h.List()) != 3 {
		t.Error("unable to create tabs")
	}

	h.Remove(1111)
	client1 = h.Create(1234)

	if !client1.Active {
		t.Error("unable to retrieve tab")
	}

	client2 = h.Create(1111)
	if client2.Active {
		t.Error("remove failed, client maintained state across re-creation")
	}

	h.Remove(1111)
	
	if len(h.List()) != 2 {
		t.Error("unable to remove item from list")
	}
}
