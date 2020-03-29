package build

import (
	"testing"

	"github.com/altid/libs/config/internal/entry"
)

func TestPush(t *testing.T) {
	test := &struct {
		Name string `altid:"Some string"`
		Port int    `altid:",omit_empty"`
	}{"None", 0}

	entry := &entry.Entry{
		Key:   "Name",
		Value: "none",
	}

	if e := push(test, entry); e != nil {
		t.Error(e)
	}
}
