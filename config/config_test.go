package config

import (
	"io"
	"testing"
)

func buildConfig(rwc io.ReadWriter) (*Config, error) {
	c := &Config{
		Name: "zzyzx",
		Values: []*Entry{
			&Entry{
				Key:   "address",
				Value: "123.456.789.0",
			},
		},
	}

	return c, nil
}

func TestCreate(t *testing.T) {
	conf, err := Mock(buildConfig, "zzyzx")
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

	return Repl(rw, repl)
}

func TestRepl(t *testing.T) {
	if _, e := Mock(buildReplConfig, "zzyzx"); e != nil {
		log.Fatal(e)
	}
}
