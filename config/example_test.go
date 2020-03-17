package main

import (
	"io"
	"log"

	"github.com/altid/libs/config"
)

func buildConfig(rw io.ReadWriter) (*config.Config, error) {
	repl := struct {
		Address  string `IP address to dial`
		Password string `password for service`
		UseSSL   bool   `Use SSL?`
		Foo      string // Will use default
	}{"127.0.0.1", "password", false, "banana"}

	return config.Repl(rw, repl)
}

func main() {
	if _, e := config.New(buildConfig, "zzyzx"); e != nil {
		log.Fatal(e)
	}
}
