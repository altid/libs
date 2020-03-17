package config_test

import (
	"io"
	"log"

	"github.com/altid/libs/config"
)

func ExampleRepl() {
	buildConfig := func(rw io.ReadWriter) (*config.Config, error) {
		repl := struct {
			Address  string `IP address to dial`
			Password string `password for service`
			UseSSL   bool   `Use SSL?`
			Foo      string // Will use default
		}{"127.0.0.1", "password", false, "bar"}

		return config.Repl(rw, repl, false)
	}

	if _, e := config.Mock(buildConfig, "zzyzx", false); e != nil {
		log.Fatal(e)
	}
}
