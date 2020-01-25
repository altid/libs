package main

import (
	"bufio"
	"bytes"
	"errors"
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
	data []byte
	size int64
}

func init() {
	s := &fileHandler{
		fn:   getTabs,
		stat: getTabsStat,
	}
	addFileHandler("/tabs", s)
}

func (t *tabs) ReadAt(p []byte, off int64) (n int, err error) {
	fmt.Println("in call")
	n = copy(p, t.data[:off])
	if int64(n)+off > t.size {
		return n, io.EOF
	}

	return
}

func (t *tabs) WriteAt(p []byte, off int64) (int, error) {
	return 0, errors.New("writes not allowed to feed")
}

func (t *tabs) Close() error { return nil }
func (t *tabs) Uid() string  { return defaultUID }
func (t *tabs) Gid() string  { return defaultGID }

func getTabs(msg *message) (interface{}, error) {
	var b bytes.Buffer
	for name, tab := range msg.svc.tabs {
		fmt.Fprintf(&b, "%s [%d]\n", name, tab.count)
	}
	t := tabs{
		size: int64(b.Len()),
		data: b.Bytes(),
	}

	return t, nil
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
