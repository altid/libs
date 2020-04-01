package conf

import (
	"log"
	"os"
	"testing"

	"github.com/altid/libs/config/internal/entry"
	"github.com/altid/libs/config/internal/request"
	"github.com/altid/libs/config/types"
)

func TestCreate(t *testing.T) {
	test := &struct {
		// Ideally, we would test the enum on auth as well
		// But without stdin/stdout rigged up for testing
		// this is not possible. Manual testing occurs in lieu
		// of automated testing for this specific instance
		Auth   types.Auth `altid:"auth,no_prompt"`
		Log    types.Logdir
		Listen types.ListenAddress
		Name   string
		Port   int32
		Delta  float32
		Debug  bool
	}{"password", "none", "none", "halfwit", 1234, 1.43424321, true}

	reqs, err := request.Build(test)
	if err != nil {
		t.Error(err)
	}

	l := log.New(os.Stdout, "conf test: ", 0)

	var entries []*entry.Entry
	for _, req := range reqs {
		entry, err := fillEntry(l.Printf, req)
		if err != nil {
			t.Error(err)
			return
		}

		entries = append(entries, entry)
	}

	entries = append(entries, &entry.Entry{
		Key:   "password",
		Value: "hunter2",
	})

	conf := &Conf{
		name:    "test",
		entries: entries,
	}

	if e := conf.WriteOut(); e != nil {
		t.Error(err)
	}
}
