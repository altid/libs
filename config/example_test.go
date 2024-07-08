package config_test

import (
	"log"

	"github.com/altid/libs/config"
)

func ExampleMarshal() {
	conf := struct {
		Address string     `altid:"address,prompt:IP address to dial"`
		UseSSL  bool       `altid:"usessl,prompt:Use SSL?,pick:true|false"`
		Foo     string     // Will use default
	}{"127.0.0.1", false, "bar"}

	if e := config.Marshal(&conf, "zzyzx", "resources/marshal_config", false); e != nil {
		log.Fatal(e)
	}
}

func ExampleCreate() {
	conf := struct {
		Address string              `altid:"address,prompt:IP to dial"`
		Port    int                 `altid:"port,no_prompt"`
	}{"irc.freenode.net", 1234}

	if e := config.Create(&conf, "zzyzx", "resources/create_config", true); e != nil {
		log.Fatal(e)
	}
}
