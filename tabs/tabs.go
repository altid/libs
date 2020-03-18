package tabs

import (
	"bytes"
	"fmt"
)

// Track the internal tab state
const (
	Alert uint8 = 1 << iota
	Active
	ClearUnread
)

// Manager is used to manage tabs accurately for a service
type Manager struct {
	tabs []*Tab
}

// Tab represents the state of one open buffer
type Tab struct {
	Name   string
	Alert  bool
	Active bool
	Unread uint16
}

// List returns all currently tracked tabs
func (m *Manager) List() []*Tab {
	return m.tabs
}

// Push

// Pop

// SetState will set Alert and Active flags of a tag based on the mask
func (t *Tab) SetState(mask uint8) {
	t.Alert = false

	switch {
	case ClearUnread&mask != 0:
		t.Unread = 0
		fallthrough
	case Alert&mask != 0:
		t.Alert = true
		fallthrough
	case Active&mask != 0:
		t.Active = true
	default:
		t.Active = false
	}
}

func (t *Tab) String() string {
	var b bytes.Buffer

	if t.Alert {
		b.WriteRune('!')
	}

	fmt.Fprintf(&b, "[%d] ", t.Unread)
	b.WriteString(t.Name)

	return b.String()
}
