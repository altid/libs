package tabs

import (
	"bytes"
	"fmt"
)

// Tab represents the state of one open buffer
type Tab struct {
	alert  bool
	unread uint16
	active bool
	name   string
	refs   uint
}

func (t *Tab) IsActive() bool {
	return t.active
}

// Activity updates the tab
// if tab is active, it will clear any remaining unreads/alerts
// if inactive, it will increment the unread count
func (t *Tab) Activity() {
	switch t.active {
	case false:
		t.unread++
	case true:
		t.unread = 0
		t.alert = false
	}
}

func (t *Tab) Alert() {
	t.alert = true
}

func (t *Tab) String() string {
	var b bytes.Buffer

	if t.alert {
		b.WriteRune('!')
	}

	fmt.Fprintf(&b, "[%d] ", t.unread)
	b.WriteString(t.name)

	return b.String()
}
