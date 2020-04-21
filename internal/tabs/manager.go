package tabs

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path"
	"sync"
)

// Manager is used to manage tabs accurately for a service
type Manager struct {
	tabs map[string]*Tab
	sync.Mutex
}

// FromFile returns a Manager with all tabs listed in a file added
func FromFile(dir string) (*Manager, error) {
	t := &Manager{
		tabs: make(map[string]*Tab),
	}

	fp, err := os.Open(path.Join(dir, "tabs"))
	if err != nil {
		return nil, err
	}

	defer fp.Close()

	s := bufio.NewReader(fp)
	for {
		line, err := s.ReadString('\n')
		switch err {
		case io.EOF:
			// We may have a tab here as well, add it
			if len(line) > 0 {
				t.Tab(line[:len(line)-1])
			}

			if len(t.List()) > 0 {
				return t, nil
			}

			return nil, errors.New("found no entries")
		case nil:
			t.Tab(line[:len(line)-1])
		default:
			return nil, err
		}

	}
}

// Default returns the first tab from the list
// or none if no tabs exist
func (m *Manager) Default() string {
	for _, t := range m.tabs {
		return t.name
	}

	return "none"
}

// List returns all currently tracked tabs
func (m *Manager) List() []*Tab {
	var tabs []*Tab
	for _, tab := range m.tabs {
		tabs = append(tabs, tab)
	}

	return tabs
}

// Tab returns a named tab, creating and appending to our list if none exists
func (m *Manager) Tab(name string) *Tab {
	tab, ok := m.tabs[name]
	if !ok {
		tab = &Tab{
			name: name,
		}
		m.tabs[name] = tab
	}

	return tab
}

// Remove a named tab from the internal list
func (m *Manager) Remove(name string) error {
	delete(m.tabs, name)
	return nil
}

// Done decrements the count on a named tab
// if the tab reference count is 0, mark it as inactive
func (m *Manager) Done(name string) {
	tab, ok := m.tabs[name]
	if !ok {
		return
	}

	tab.refs--
	if tab.refs < 1 {
		tab.refs = 0
		tab.active = false
	}

	// When we're done with a tab, we can also zero it
	tab.unread = 0
}

// Active incements the count on a named tab
func (m *Manager) Active(name string) {
	tab, ok := m.tabs[name]
	if !ok {
		return
	}

	tab.refs++
	tab.active = true
}
