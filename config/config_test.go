package config

import (
	"testing"
)

func TestMarshal(t *testing.T) {
	conf := struct {
		Address string     `altid:"address,prompt:IP address to dial"`
		UseSSL  bool       `altid:"usessl,prompt:use ssl?"`
		Foo     string     // Will use default
	}{"127.0.0.1", false, "bar"}

	if e := Marshal(&conf, "zzyzx", "resources/marshal_config", true); e != nil {
		t.Error(e)
	}

	if conf.Address != "test" {
		t.Error("unable to set address field in conf")
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
		Logdir     Logdir
		Listen  ListenAddress
	}{"irc.freenode.net", 1234, "", ""}

	if e := Create(&conf, "zzyzx", "resources/create_config", true); e != nil {
		t.Error(e)
	}
}
*/
