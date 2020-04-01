package config_test

import (
	"log"

	"github.com/altid/libs/config"
	"github.com/altid/libs/config/types"
)

func ExampleMarshal() {
	conf := struct {
		Address string     `altid:"address,prompt:IP address to dial"`
		Auth    types.Auth `altid:"auth,prompt:Auth mechanism to use,pick:password|factotum|none"`
		UseSSL  bool       `altid:"usessl,prompt:Use SSL?,pick:true|false"`
		Foo     string     // Will use default
	}{"127.0.0.1", "password", false, "bar"}

	if e := config.Marshal(&conf, "zzyzx", "resources/marshal_config", false); e != nil {
		log.Fatal(e)
	}
}

func ExampleCreate() {
	conf := struct {
		Address string              `altid:"address,prompt:IP to dial"`
		Port    int                 `altid:"port,no_prompt"`
		Auth    types.Auth          `altid:"auth,prompt:Auth mechanism to use,pick:password|factotum|none"`
		Logdir  types.Logdir        `altid:"logdir,no_prompt"`
		Listen  types.ListenAddress `altid:"listen_address,no_prompt"`
	}{"irc.freenode.net", 1234, "none", "", ""}

	if e := config.Create(&conf, "zzyzx", "resources/create_config", true); e != nil {
		log.Fatal(e)
	}
}
