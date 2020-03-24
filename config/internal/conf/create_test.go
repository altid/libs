package conf

import (
	"log"
	"os"
	"testing"

	"github.com/altid/libs/config/internal/request"
	"github.com/altid/libs/config/types"
)

func TestCreate(t *testing.T) {
	test := &struct {
		Auth   types.Auth
		Log    types.Logdir
		Listen types.ListenAddress
		Name   string
		Port   int32
		Delta  float32
		Debug  bool
	}{"none", "none", "none", "halfwit", 1234, 1.4, true}

	reqs, err := request.Build(test)
	if err != nil {
		t.Error(err)
	}

	l := log.New(os.Stdout, "conf test: ", 0)
	
	for _, req := range reqs {
		_, err := fillEntry(l.Printf, req)
		if err != nil {
			t.Error(err)
		}
	}
}
