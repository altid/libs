package config_test

import (
	"crypto/tls"
	"log"

	"github.com/altid/libs/config"
)

func ExampleMarshall() {
	conf := struct {
		Address  string `IP address to dial`
		Auth config.Auth `Auth mechanism to use: password|factotum|none`
		UseSSL   bool   `Use SSL?`
		Foo      string // Will use default
	}{"127.0.0.1", "password", false, "bar"}

	if e := config.Marshal(&conf, "zzyzx", "resources/marshall_config", false); e != nil {
		log.Fatal(e)
	}
}

func ExampleCreate() {
	conf := struct {
		Address string `Enter the address you wish to connect on`
		Port    int
		Auth    config.Auth `Enter your authentication method: password|factotum|none`
		Logdir     config.Logdir
		Tls     tls.Certificate
		Listen  config.ListenAddress
	}{"irc.freenode.net", 1234, "none", "", tls.Certificate{}, ""}

	if e := config.Create(&conf, "zzyzx", "resources/create_config", true); e != nil {
		log.Fatal(e)
	}
}
