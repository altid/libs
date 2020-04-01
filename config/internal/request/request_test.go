package request

import (
	"fmt"
	"testing"

	"github.com/altid/libs/config/types"
)

func TestRequest(t *testing.T) {
	// Create and marshal a struct, make sure we're good here
	test := &struct {
		Auth   types.Auth          `altid:"auth,prompt:Authentication method,pick:password|factotum|none"`
		Log    types.Logdir        `altid:"logdir,no_prompt"`
		Listen types.ListenAddress `altid:"listen_address,no_prompt"`
		Name   string              `altid:"username,prompt:Name to use for service"`
		Port   int32               `altid:"port,prompt:Port to connect on"`
		Debug  bool                `xml:"debug,omit_empty"`
	}{"password", "none", "none", "halfwit", 1234, true}

	reqs, err := Build(test)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 5; i++ {
		fmt.Println(reqs[i])
	}

	if len(reqs) != 5 {
		t.Error("incorrect number of requests found")
	}

	if len(reqs[0].Pick) != 3 {
		t.Error("unable to find picks for auth")
	}

	if reqs[0].Prompt == "" {
		t.Error("was unable to set prompt for Auth")
	}

	if reqs[4].Defaults.(int32) != 1234 {
		t.Error("was unable to set int type")
	}
}
