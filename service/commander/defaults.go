package commander

// TODO(halfiwt) i18n
var DefaultCommands = []*Command{
	{
		Name:        "open",
		Args:        []string{"<buffer>"},
		Heading:     DefaultGroup,
		Description: "Open and change buffers to a given service",
	},
	{
		Name:        "close",
		Args:        []string{"<buffer>"},
		Heading:     DefaultGroup,
		Description: "Close a buffer and return to the last opened previously",
	},
	{
		Name:        "buffer",
		Args:        []string{"<buffer>"},
		Heading:     DefaultGroup,
		Description: "Change to the named buffer",
	},
	{
		Name:        "link",
		Args:        []string{"<current>", "<buffer>"},
		Heading:     DefaultGroup,
		Description: "Overwrite the current buffer with the named",
	},
	{
		Name:        "quit",
		Args:        []string{},
		Heading:     DefaultGroup,
		Description: "Exits the client",
	},
}
