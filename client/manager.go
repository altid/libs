package client

import (
	"errors"
	"sync"
)

// Manager is used to represent an overview of all currently
// connected clients
type Manager struct {
	clients []*Client
	sync.Mutex
}

// List returns all currently tracked clients
func (m *Manager) List() []*Client {
	return m.clients
}

// Create a client for id, and append to our manager list with buffer "none", if it does not exist
func (m *Manager) Create(uuid uint32) *Client {
	for _, c := range m.clients {
		if c.uuid == uuid {
			return c
		}
	}

	client := &Client{
		uuid:    uuid,
		current: "none",
	}

	m.Lock()
	m.clients = append(m.clients, client)
	m.Unlock()

	return client
}

// Remove a named tab from the internal list
func (m *Manager) Remove(uuid uint32) error {
	for n, c := range m.clients {
		if c.uuid == uuid {
			m.Lock()
			m.clients[n] = m.clients[len(m.clients)-1]
			m.clients = m.clients[:len(m.clients)-1]
			m.Unlock()

			return nil
		}
	}
	return errors.New("No client found")
}
