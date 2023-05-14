package config

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/altid/libs/config/internal/conf"
	"github.com/altid/libs/config/internal/entry"
	"github.com/altid/libs/config/internal/util"
	"github.com/altid/libs/service"
	"github.com/mischief/ndb"
)

// Marshal will take a pointer to a struct as input, as well as the name of the service and attempt to fill the struct.
// The struct tags with a matching altid prefix will be marshalled, according to the defaults set
// The first value of the tag must be the Altid config name you wish to retrieve, such as listen_address
// Other fields include prompt/no_prompt:, and pick:
// Pick will return an error if a value other than one listed exists in the config/defaults
//
//	conf := struct {
//		Address string     `altid:"address,prompt:IP address to dial"`
//		Auth    types.Auth `altid:"auth,prompt:Auth mechanism to use,pick:password|factotum|none"`
//		UseSSL  bool       `altid:"usessl,prompt:Use SSL?,pick:true|false"`
//		Foo     string     // Will use default
//	}{"127.0.0.1", "password", false, "bar"}
//
//	if e := config.Marshal(&conf, "zzyzx", "", false); e != nil {
//		log.Fatal(e)
//	}
//  [...]
//
func Marshal(requested interface{}, service string, configFile string, debug bool) error {
	debugLog := func(string, ...interface{}) {}
	if debug {
		debugLog = func(format string, v ...interface{}) {
			l := log.New(os.Stdout, "config: ", 0)
			l.Printf(format+"\n", v...)
		}
	}

	// list all existing config entries
	have, err := entry.FromConfig(debugLog, service, configFile)
	if err != nil {
		return err
	}

	if e := conf.Marshal(debugLog, requested, have, nil); e != nil {
		return e
	}

	return nil
}

// Create takes a pointer to a struct, and will walk a user through creation of a config entry for the service
// Upon success, it will print the config and instructions to stdout
// It is meant to be used with the -conf flag
//
//	conf := struct {
//		Address string     `altid:"address,prompt:IP address to dial"`
//		Auth    types.Auth `altid:"auth,prompt:Auth mechanism to use"`
//		UseSSL  bool       `altid:"usessl,prompt:Use SSL?,pick:true|false"`
//		Foo     string     // Will use default
//	}{"127.0.0.1", "password", false, "bar"}
//
//	if e := config.Create(&conf, "zzyzx", "", false); e != nil {
//		log.Fatal(e)
//	}
//
//  os.Exit(0)
//
// Notably, Create will parse entries for altid struct tags with the field "prompt". These will prompt a user
// for the value on the command line, optionally with a whitelisted array of selections to pick from
// Selection of an item not on a whitelist will return an error after 3 attempts
// The `pick` option to a types.Auth will be ignored, and will always be one of `password|factotum|none`
func Create(requests interface{}, svc, configFile string, debug bool) error {
	debugLog := func(string, ...interface{}) {}
	if debug {
		debugLog = func(format string, v ...interface{}) {
			l := log.New(os.Stdout, "config: ", 0)
			l.Printf(format+"\n", v...)
		}
	}

	have, err := entry.FromConfig(debugLog, svc, configFile)
	// Make sure we correct any errors we encounter
	switch {
	case err == nil:
		debugLog("no errors in config")
	case os.IsNotExist(err):
		dir, _ := service.UserConfDir()
		os.MkdirAll(path.Join(dir, "altid"), 0755)
		os.Create(util.GetConf(""))
		debugLog("creating config file")
	// If we have multiple entries, something has indeed gone wrong
	// The user needs to manually clean this up
	case err.Error() == entry.ErrMultiEntries:
		return err
	// This is the expected case in this situation
	case err.Error() == entry.ErrNoEntries:
		debugLog("creating entry")
	default:
		debugLog("error: %v\n", err)
	}

	if e := conf.Marshal(debugLog, requests, have, conf.NewPrompt(debugLog)); e != nil {
		return e
	}

	return conf.WriteOut(svc, requests)
}

// GetListenAddress returns the listen_address of a server, or "" if none is found
// If a port is set, e.g. listen_address = 192.168.0.4:8080 it will return 8080
func GetListenAddress(service string) (string, string) {
	conf, err := ndb.Open(util.GetConf(service))
	if err != nil {
		return "", "564"
	}

	listen := conf.Search("service", service).Search("listen_address")
	if listen != "" {
		if n := strings.Index(listen, ":"); n > 0 {
			return listen[:n], listen[n+1:]
		}

		return listen, "564"
	}

	return "", "564"
}

// GetLogDir returns a canonical directory for a user log, searching first altid/config
// If no entry is found or the file is missing, it will return "none"
func GetLogDir(service string) string {
	conf, err := ndb.Open(util.GetConf(service))
	if err != nil {
		return "none"
	}

	logdir := conf.Search("service", service).Search("log")
	if logdir != "" {
		return path.Join(logdir, service)
	}

	return "none"
}

// ListAll returns a list of available services
func ListAll() ([]string, error) {
	var configs []string

	conf, err := ndb.Open(util.GetConf(""))
	if err != nil {
		return nil, err
	}

	for _, rec := range conf.Search("service", "") {
		configs = append(configs, rec[0].Val)
	}

	return configs, nil
}
