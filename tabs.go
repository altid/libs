package main

import (
	"bufio"
	"os"
	"path"
	"strings"
)

// tabs are a special file type that must track unread across all clients with a state
type tab struct {
	count  uint16
	active bool
}

func init() {
	s := &fileHandler{
		fn:   getTabs,
		stat: getTabsStat,
	}
	addFileHandler("/tabs", s)
}

func getTabs(msg *message) (interface{}, error) {
	return nil, nil
}

func getTabsStat(msg *message) (os.FileInfo, error) {
	return nil, nil
}

func listInitialTabs(service string) (map[string]*tab, error) {
	var tabs map[string]*tab
	tabs = make(map[string]*tab)
	fp := path.Join(*inpath, service, "tabs")
	file, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	r := bufio.NewReader(file)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return tabs, nil
		}
		name := strings.TrimSpace(line)
		tabs[name] = &tab{}
	}
}
