package fs

import (
	"context"
	"errors"
	"fmt"
)

type tab struct {
	name    string
	doctype string
	data    []byte
}

type mockctl struct {
	reqs chan string
	cmds chan string
	done chan struct{}
	err  chan error
	tabs []*tab
}

func (t *tab) Write(p []byte) (n int, err error) {
	n = copy(p, t.data)
	return
}

func (t *tab) Close() error {
	return nil
}

func (tc *mockctl) cleanup() {}

func (tc *mockctl) event(ev string) error {
	return nil
}

func (tc *mockctl) createBuffer(name, doctype string) error {
	return tc.pushTab(name, doctype)
}

func (tc *mockctl) deleteBuffer(name, doctype string) error {
	return tc.popTab(name)
}

func (tc *mockctl) hasBuffer(name, doctype string) bool {
	for _, i := range tc.tabs {
		if i.name == name {
			return true
		}
	}

	return false
}

func (tc *mockctl) remove(name, doctype string) error {
	return tc.popTab(name)
}

func (tc *mockctl) listen() error {
	defer close(tc.err)
	defer close(tc.done)

	for {
		select {
		case cmd := <-tc.reqs:
			if cmd == "quit" {
				return nil
			}
			tc.cmds <- cmd
		case err := <-tc.err:
			return err
		}
	}
}

func (tc *mockctl) start() (context.Context, error) {
	return nil, errors.New("please use listen for testing")
}

func (tc *mockctl) notification(string, string, string) error {
	return nil
}

func (tc *mockctl) popTab(tabname string) error {
	for n := range tc.tabs {
		if tc.tabs[n].name == tabname {
			tc.tabs = append(tc.tabs[:n], tc.tabs[n+1:]...)
			return nil
		}
	}

	return fmt.Errorf("entry not found: %s", tabname)
}

func (tc *mockctl) pushTab(tabname, doctype string) error {
	for n := range tc.tabs {
		if tc.tabs[n].name == tabname {
			return fmt.Errorf("entry already exists: %s", tabname)
		}
	}

	t := &tab{
		name:    tabname,
		doctype: doctype,
	}

	tc.tabs = append(tc.tabs, t)
	return nil
}

func (tc *mockctl) errorwriter() (*WriteCloser, error) {
	w := &WriteCloser{
		c:      tc,
		fp:     &tab{},
		buffer: "errors",
	}

	return w, nil
}
func (tc *mockctl) fileWriter(buffer, doctype string) (*WriteCloser, error) {
	tab, err := tc.findTab(buffer)
	if err != nil {
		tc.err <- err
		return nil, err
	}

	w := &WriteCloser{
		c:      tc,
		fp:     tab,
		buffer: buffer,
	}

	return w, nil
}
func (tc *mockctl) imageWriter(buffer, resource string) (*WriteCloser, error) {
	tab, err := tc.findTab(buffer)
	if err != nil {
		tc.err <- err
		return nil, err
	}

	w := &WriteCloser{
		c:      tc,
		fp:     tab,
		buffer: buffer,
	}

	return w, nil
}

func (tc *mockctl) findTab(buffer string) (*tab, error) {
	for _, tab := range tc.tabs {
		if tab.name == buffer {
			return tab, nil
		}
	}

	return nil, errors.New("No such tab")
}
