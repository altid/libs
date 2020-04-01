package entry

import (
	"testing"

	"github.com/altid/libs/config/internal/request"
)

func TestFind(t *testing.T) {
	r := &request.Request{
		Field:   "Auth",
		Key:     "auth",
		Prompt:  "Auth to use",
		Defaults: "password",
	}

	entries := []*Entry{
		{
			Key:   "auth",
			Value: "password",
		},
	}

	if _, ok := Find(r, entries); ! ok {
		t.Error("Find error: could not locate entry")
	}
}
