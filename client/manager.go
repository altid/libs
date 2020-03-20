package client

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

// UUID is a unique identifier for a client
type UUID uint32

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

// Client - return for given id
// If UUID is 0, a new one will be generated
func (m *Manager) Client(id UUID) *Client {

	for _, c := range m.clients {
		if c.UUID == id {
			return c
		}
	}

	if id == 0 {
		newid := uuid.New()
		id = UUID(newid.ID())
	}

	client := &Client{
		UUID:    id,
		current: "none",
	}

	m.Lock()
	m.clients = append(m.clients, client)
	m.Unlock()

	return client
}

// Remove a named tab from the internal list
func (m *Manager) Remove(uuid UUID) error {
	for n, c := range m.clients {
		if c.UUID == uuid {
			m.Lock()
			m.clients[n] = m.clients[len(m.clients)-1]
			m.clients = m.clients[:len(m.clients)-1]
			m.Unlock()

			return nil
		}
	}
	return errors.New("No client found")
}
