package fs

import (
	"context"
	"fmt"
)

type tab struct {
	name string
	doct string
	data string
}

type mockctl struct {
	req  chan string
	done chan struct{}
	tabs []*tab
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
	for {
		// Wait forever
	}

	return nil
}

func (tc *mockctl) start() (context.Context, error) {
	return context.Background(), nil
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
		name: tabname,
		doct: doctype,
	}

	tc.tabs = append(tc.tabs, t)

	return nil
}

func (tc *mockctl) errorwriter() (*WriteCloser, error) {
	return nil, nil
}
func (tc *mockctl) fileWriter(buffer, doctype string) (*WriteCloser, error) {
	return nil, nil
}
func (tc *mockctl) imageWriter(buffer, resource string) (*WriteCloser, error) {
	return nil, nil
}
