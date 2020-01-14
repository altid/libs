package main

// tabs are a special file type that must track unread across all clients with a state
type tab struct {
	// Or so, we aren't married to this
	count  uint16
	active bool
}

var tabs map[string]*tab

func init() {
	tabs = make(map[string]*tab)
	s := &fileHandler{
		fn: getTabs,
	}
	addFileHandler("/tabs", s)
}

func getTabs(msg *message) (interface{}, error) {
	return tabs, nil
}
