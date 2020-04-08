package tabs

import (
	"fmt"
	"testing"
)

func TestManager(t *testing.T) {
	// FromFile
	m, err := FromFile("resources")
	if err != nil {
		t.Error(err)
	}

	t1 := m.Tab("5th_element")
	if len(m.List()) > 5 {
		fmt.Println(m.List())
		t.Error("Could not retreive existing tab")
	}

	t1.Alert()

	if e := m.Remove("5th_element"); e != nil {
		t.Error("unable to remove item")
	}

	if len(m.List()) != 4 {
		t.Error("unable to remove item from list")
	}

	t2 := m.Tab("5th_element")
	if t2.String()[0] == '!' {
		t.Error("Active wasn't upset after remove")
	}

	t2.Alert()

	if t2.String()[0] != '!' {
		t.Error("Active not set")
	}

	m.Active("5th_element")
	t2.Activity()

	if t2.String()[0] == '!' {
		t.Error("Activity did not clear alert")
	}

	m.Done("5th_element")
	t2.Activity()

	if t2.String()[1] != '1' {
		t.Error("Activity did not increase unread count")
	}

	t2.Activity()

	if t2.String()[1] != '2' {
		t.Error("Activity did not increase unread count")
	}

	m.Active("5th_element")
	t2.Activity()

	if t2.String()[1] != '0' {
		t.Error("Activity on active buffer did not reset unread count")
	}
}
