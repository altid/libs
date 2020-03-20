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
	tabs []*Tab
	sync.Mutex
}

func FromFile(dir string) (*Manager, error) {
	t := &Manager{}

	fp, err := os.Open(path.Join(dir, "tabs"))
	if err != nil {
		return nil, err
	}

	defer fp.Close()

	s := bufio.NewReader(fp)
	for {
		line, err := s.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return t, nil
			}
			return nil, err
		}
		t.Tab(line[:len(line)-1])
	}

	return t, nil
}

// List returns all currently tracked tabs
func (m *Manager) List() []*Tab {
	return m.tabs
}

// Tab returns a named tab, creating and appending to our list if none exists
func (m *Manager) Tab(name string) *Tab {
	for _, t := range m.tabs {
		if t.Name == name {
			return t
		}
	}

	tab := &Tab{
		Name: name,
	}

	m.Lock()
	m.tabs = append(m.tabs, tab)
	m.Unlock()

	return tab
}

// Remove a named tab from the internal list
func (m *Manager) Remove(name string) error {
	for n, t := range m.tabs {
		if t.Name == name {
			m.Lock()
			m.tabs[n] = m.tabs[len(m.tabs)-1]
			m.tabs = m.tabs[:len(m.tabs)-1]
			m.Unlock()

			return nil
		}
	}
	return errors.New("no such tab exists")
}
