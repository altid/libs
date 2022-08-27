package config

import (
	"testing"

	"github.com/altid/libs/config/types"
)

func TestMarshal(t *testing.T) {
	conf := struct {
		Address string     `altid:"address,prompt:IP address to dial"`
		Auth    types.Auth `altid:"auth,prompt:mechanism to use"`
		UseSSL  bool       `altid:"usessl,prompt:use ssl?"`
		Foo     string     // Will use default
	}{"127.0.0.1", "password", false, "bar"}

	if e := Marshal(&conf, "zzyzx", "resources/marshal_config", true); e != nil {
		t.Error(e)
	}

	if conf.Address != "test" {
		t.Error("unable to set address field in conf")
	}

	if conf.Auth != "banana" {
		t.Error("unable to set password based on auth mechanism")
	}

	if conf.UseSSL != true {
		t.Error("unable to set UseSSL boolean")
	}
}

/* Testing create requires a stdin/stdout, therefore this is audited manually
func TestCreate(t *testing.T) {
	conf := struct {
		Address string `Enter the address you wish to connect on`
		Port    int
		Auth    Auth `Enter your authentication method: password|factotum|none`
		Logdir     Logdir
		Listen  ListenAddress
	}{"irc.freenode.net", 1234, "none", "", ""}

	if e := Create(&conf, "zzyzx", "resources/create_config", true); e != nil {
		t.Error(e)
	}
}
*/
