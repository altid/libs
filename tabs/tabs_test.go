package tabs

import (
	"fmt"
	"testing"
)

func TestManager(t *testing.T) {
	h := Manager{}

	tab := h.Create("foo")
	tab.SetState(Active)

	h.Create("bar")
	tab2 := h.Create("baz")
	tab2.SetState(Alert)

	l := h.List()

	if len(l) != 3 {
		t.Error("unable to create tabs")
	}

	h.Remove("baz")
	tab = h.Create("foo")

	if !tab.Active {
		t.Error("unable to retrieve tab")
	}

	tab2 = h.Create("baz")
	if tab.Alert {
		t.Error("remove failed, tab maintained state across re-creation")
	}

	h.Remove("baz")
	if len(h.List()) != 2 {
		t.Error("unable to remove entry entirely")
	}
}

func TestTab(t *testing.T) {
	// Make sure we tax the management, change up everything we can and ensure tracking is rock solid
	d := &Tab{
		Name:   "test",
		Alert:  true,
		Unread: 342,
	}

	if fmt.Sprintf("%s", d) != "![342] test" {
		t.Error("incorrect printer for type tabs")
	}

	d.SetState(ClearUnread | Alert | Active)
	if !(d.Active && d.Alert) {
		t.Error("unable to set flags")
	}

	if d.Unread > 0 {
		t.Error("unable to clear unreads")
	}

	d.SetState(Active)

	if fmt.Sprintf("%s", d) != "[0] test" {
		t.Error("incorrect state after modification to tabs")
	}
}
