package client

// Client represents a unique client attatched to a server
type Client struct {
	Active  bool
	uuid    uint32
	current string
	history []string
}

// SetBuffer updates the client's current buffer
// if set to "none", it marks it as inactive
// if buffer was previously inactive, it marks it as active
func (c *Client) SetBuffer(buffer string) {
	// Setting to inactive
	if buffer == "none" {
		if c.current != "none" {
			c.history = append(c.history, c.current)
		}

		c.Active = false
		c.current = buffer
		return
	}

	// Coming out of inactive
	if c.current == "none" && buffer != "none" {
		c.Active = true
		c.current = buffer
		return
	}

	// Normal
	c.Active = true
	c.history = append(c.history, c.current)
	c.current = buffer
}
