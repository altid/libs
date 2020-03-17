package config

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func buildConfig(rwc io.ReadWriter) (*Config, error) {
	c := &Config{
		Name: "zzyzx",
		Values: []*Entry{
			{
				Key:   "address",
				Value: "123.456.789.0",
			},
		},
	}

	return c, nil
}

func TestCreate(t *testing.T) {
	conf, err := Mock(buildConfig, "zzyzx", true)
	if err != nil {
		t.Error(err)
	}

	if conf.Values[0].Key != "address" && conf.Values[0].Value != "123.456.789.0" {
		t.Error("unable to create config successfully")
	}
}

func buildReplConfig(rw io.ReadWriter) (*Config, error) {
	repl := struct {
		Address  string `IP address to dial`
		Password string `password for service`
		UseSSL   bool   `Use SSL?`
		Foo      string // Will use default
	}{"127.0.0.1", "password", false, "banana"}

	rw = struct {
		io.Reader
		io.Writer
	}{os.Stdin, ioutil.Discard}

	return Repl(rw, repl, true)
}

func TestRepl(t *testing.T) {
	if _, e := Mock(buildReplConfig, "zzyzx", true); e != nil {
		log.Fatal(e)
	}
}
