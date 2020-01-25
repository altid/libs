package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

// tabs are a special file type that must track unread across all clients with a state
type tab struct {
	name   string
	count  uint16
	active bool
}

type tabs struct {
	t    chan *tab
	done chan struct{}
}

func init() {
	s := &fileHandler{
		fn:   getTabs,
		stat: getTabsStat,
	}
	addFileHandler("/tabs", s)
}

func (t *tabs) ReadAt(b []byte, off int64) (n int, err error) {
	var buff bytes.Buffer

	for tab := range t.t {
		fmt.Fprintf(&buff, "%s [%d]\n", tab.name, tab.count)
	}

	n = copy(b, buff.Bytes())
	if int64(n)+off > int64(buff.Len()) {
		return n, io.EOF
	}

	return
}

func (t *tabs) Close() {
	close(t.done)
}

func getTabs(msg *message) (interface{}, error) {
	t := make(chan *tab)
	done := make(chan struct{})

	go func(msg *message, t chan *tab, done chan struct{}) {
		for name, tab := range msg.svc.tabs {
			tab.name = name
			select {
			case t <- tab:
				continue
			case <-done:
				break
			}
		}
	}(msg, t, done)

	b := &tabs{
		t:    t,
		done: done,
	}

	return b, nil
}

func getTabsStat(msg *message) (os.FileInfo, error) {
	return os.Lstat(path.Join(*inpath, msg.svc.name, "tabs"))
}

func listInitialTabs(service string) (map[string]*tab, error) {
	tabs := make(map[string]*tab)
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
