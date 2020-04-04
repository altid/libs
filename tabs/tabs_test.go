package tabs

import (
	"fmt"
	"testing"
)

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
