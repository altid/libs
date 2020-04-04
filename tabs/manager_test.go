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

	// Push
	t1 := m.Tab("5th_element")
	if len(m.List()) > 5 {
		fmt.Println(m.List())
		t.Error("Could not retreive existing tab")
	}

	t1.Alert = true

	// Pop
	if e := m.Remove("5th_element"); e != nil {
		t.Error("unable to remove item")
	}

	if len(m.List()) != 4 {
		t.Error("unable to remove item from list")
	}

	t2 := m.Tab("5th_element")
	if t2.Alert == true {
		t.Error("Unable to remove and clear tab")
	}
}
