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
type tabs struct {
	name   string
	alert  bool
	count  uint16
	active bool
}

type tabfile struct {
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

func (t *tabfile) ReadAt(p []byte, off int64) (n int, err error) {
	n = copy(p, t.data[off:])
	if int64(n)+off > t.size {
		return n, io.EOF
	}

	return
}

func (t *tabfile) WriteAt(p []byte, off int64) (int, error) {
	return 0, errors.New("tabs file does not allow modification")
}

func getTabs(msg *message) (interface{}, error) {
	var b bytes.Buffer

	for name, tab := range msg.svc.tablist {
		if tab.alert {
			b.WriteString("!")
		}
		
		fmt.Fprintf(&b, "%s [%d]\n", name, tab.count)
	}

	t := &tabfile{
		size: int64(b.Len()),
		data: b.Bytes(),
	}

	return t, nil
}

func getTabsStat(msg *message) (os.FileInfo, error) {
	return os.Stat(path.Join(*inpath, msg.svc.name, "tabs"))
}

func listInitialTabs(service string) (map[string]*tabs, error) {
	tablist := make(map[string]*tabs)
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
			return tablist, nil
		}

		name := strings.TrimSpace(line)
		tablist[name] = &tabs{}
	}
}
