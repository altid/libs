package build

import (
	"testing"

	"github.com/altid/libs/config/internal/entry"
	"github.com/altid/libs/config/internal/request"
	"github.com/altid/libs/config/types"
)

func TestPush(t *testing.T) {
	test := &struct {
		Name   string     `altid:"name,prompt=Some string,pick=foo|bar|baz"`
		Port   int        `altid:"chicken,omit_empty"`
		Auth   types.Auth `altid:"auth,pick=password|factotum|none"`
		Banana bool
	}{"foo", 0, "password", false}

	entry := &entry.Entry{
		Key:   "name",
		Value: "bar",
	}

	req := &request.Request{
		Pick: []string{"foo", "bar", "baz"},
	}

	if e := pick(req, entry); e != nil {
		t.Error(e)
	}

	req.Pick = []string{}

	if e := pick(req, entry); e != nil {
		t.Error(e)
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
