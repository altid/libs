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

// This is a test controller
type testController struct {
	req  chan string
	done chan struct{}
	tabs []*tab
}

func (tc *testController) cleanup() {}

func (tc *testController) event(ev string) error {
	return nil
}

func (tc *testController) createBuffer(name, doctype string) error {
	return tc.pushTab(name, doctype)
}

func (tc *testController) deleteBuffer(name, doctype string) error {
	return tc.popTab(name)
}

func (tc *testController) hasBuffer(name, doctype string) bool {
	for _, i := range tc.tabs {
		if i.name == name {
			return true
		}
	}

	return false
}

func (tc *testController) remove(name, doctype string) error {
	return tc.popTab(name)
}

func (tc *testController) listen() error {
	return nil
}

func (tc *testController) start() (context.Context, error) {
	// Need to do stuff with Open, Close, Link and Default.

	return nil, nil
}

func (tc *testController) notification(string, string, string) error {
	return nil
}

func (tc *testController) popTab(tabname string) error {
	for n := range tc.tabs {
		if tc.tabs[n].name == tabname {
			tc.tabs = append(tc.tabs[:n], tc.tabs[n+1:]...)
			return nil
		}
	}
	return fmt.Errorf("entry not found: %s", tabname)
}

func (tc *testController) pushTab(tabname, doctype string) error {
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

func (tc *testController) errorwriter() (*WriteCloser, error) {
	return nil, nil
}
func (tc *testController) fileWriter(buffer, doctype string) (*WriteCloser, error) {
	return nil, nil
}
func (tc *testController) imageWriter(buffer, resource string) (*WriteCloser, error) {
	return nil, nil
}
