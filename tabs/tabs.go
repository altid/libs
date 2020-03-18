package tabs

// Track the internal tab state entirely here, no where else
// We'll have to pass some state around in the msg from server <--> service <--> tabs

// Manager is used to manage tabs accurately for a service
type Manager struct {
	tabs []*Tab
}

// Tab represents the state of one open buffer
type Tab struct {
	Name   string
	Alert  bool
	Count  uint16
	Active bool
}

func (m *Manager) List() []*Tab {
	return m.tabs
}

// Push

// Pop

// use a const for these or a bitmask or whatever
// SetState() 
// Alert

// Clear(alert)

// Active