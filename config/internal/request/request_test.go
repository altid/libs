package request

import (
	"testing"

	"github.com/altid/libs/config/types"
)

func TestRequest(t *testing.T) {
	// Create and marshal a struct, make sure we're good here
	test := &struct {
		Auth   types.Auth `A tag`
		Log    types.Logdir
		Listen types.ListenAddress
		Name   string `Another tag`
		Port   int32
		Debug  bool
	}{"none", "none", "none", "halfwit", 1234, true}

	reqs, err := Build(test)
	if err != nil {
		t.Error(err)
	}

	if len(reqs) != 6 {
		t.Error("incorrect number of requests found")
	}

	if reqs[0].Prompt != "A tag" {
		t.Error("was unable to set prompt for Auth")
	}

	if reqs[4].Defaults.(int32) != 1234 {
		t.Error("was unable to set int type")
	}

	if reqs[5].Defaults.(bool) != true {
		t.Error("was unable to set bool type")
	}
}
