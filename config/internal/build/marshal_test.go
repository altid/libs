package build

import (
	"testing"

	"github.com/altid/libs/config/internal/entry"
)

func TestPush(t *testing.T) {
	test := &struct {
		Name string `altid:"name,prompt=Some string,pick=foo|bar|baz"`
		Port int    `altid:"chicken,omit_empty"`
		Banana bool
	}{"foo", 0, false}

	entry := &entry.Entry{
		Key:   "name",
		Value: "bar",
	}

	if e := push(test, entry); e != nil {
		t.Error(e)
	}

	if test.Name != "bar" {
		t.Error("was unable to set entry value")
	}

	entry.Key = "chicken"
	entry.Value = 26

	if e := push(test, entry); e != nil {
		t.Error(e)
	}

	if test.Port != 26 {
		t.Error("was unable to set entry value")
	}
}
