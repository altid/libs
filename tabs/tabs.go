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

// Tab represents the state of one open buffer
type Tab struct {
	Name   string
	Alert  bool
	Active bool
	Unread uint16
}

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
