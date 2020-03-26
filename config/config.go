package config

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/altid/libs/config/internal/build"
	"github.com/altid/libs/config/internal/conf"
	"github.com/altid/libs/config/internal/entry"
	"github.com/altid/libs/config/internal/request"
	"github.com/altid/libs/config/internal/util"
	"github.com/altid/libs/fs"
	"github.com/mischief/ndb"
)

// Marshal will take a pointer to a struct as input, as well as the name of the service and attempt to fill the struct.
// - The struct entries must be of the type string, int, bool,
// - The tags of the struct will be used to indicate a query sent to the user
// - Default entries to the struct will be used as defaults
//
//	type myconf struct {
//		Name string `Username to use for the service`
// 		Port int `Port to connect with`
// 		UseSSL bool `Do you want to connect with SSL?`
// 		Auth types.Auth `Auth mechanism to use: password|factotum|none`
//		Logdir types.Logdir
//      Address types.ListenAddress
//	}{myusername, 1234, true, "none", "none"}
//
//  err := config.Marshal(myconf, "myservice", false)
//  [...]
//
// The preceding example would search the config file for each lower case entry
// If it cannot fill an entry, it returns an error
// Idiomatically, the user should be prompted to rerun with the -conf flag
// and the function Create should be called, and on success, exit the program
func Marshal(requests interface{}, service string, confdir string, debug bool) error {
	debugLog := func(string, ...interface{}) {}
	if debug {
		debugLog = func(format string, v ...interface{}) {
			l := log.New(os.Stdout, "config: ", 0)
			l.Printf(format+"\n", v...)
		}
	}

	// list all existing config entries
	have, err := conf.FromConfig(debugLog, service, confdir)
	if err != nil {
		return err
	}

	if e := build.Marshal(debugLog, requests, have); e != nil {
		return e
	}

	return nil
}

// Create takes a pointer to a struct, and will walk a user through creation of a config entry for the service
// Upon success, it will print the config and instructions to stdout 
// It is meant to be used with the -conf flag
// The semantics are the same as Marshall, but it uses the struct tags to prompt the user to fill in the data for any missing entries
// For example:
//
// 	Name string `Username to connect with`
//
// would prompt the user for a username, optionally offering the default value passed in
// On success, the user should cleanly exit the program, as requests is not filled as it is in Marshall
func Create(requests interface{}, service, confdir string, debug bool) error {
	debugLog := func(string, ...interface{}) {}
	if debug {
		debugLog = func(format string, v ...interface{}) {
			l := log.New(os.Stdout, "config: ", 0)
			l.Printf(format+"\n", v...)
		}
	}

	have, err := entry.FromConfig(debugLog, service, confdir)

	// Make sure we correct any errors we encounter
	switch {
	case err == nil:
		debugLog("fixing config file")
	case os.IsNotExist(err):
		dir, err := fs.UserConfDir()
		if err != nil {
			return err
		}

		os.MkdirAll(path.Join(dir, "altid"), 0755)
		os.Create(util.GetConf(service))
		debugLog("creating config file")

	// If we have multiple entries, something has indeed gone wrong
	// The user needs to manually clean this up
	case err.Error() == entry.ErrMultiEntries:
		return err

	// This is the expected case in this situation
	case err.Error() == entry.ErrNoEntries:
		debugLog("creating entry")
	}

	want, err := request.Build(requests)
	if err != nil {
		return err
	}

	c, err := conf.Create(debugLog, service, have, want)
	if err != nil {
		return err
	}

	return c.WriteOut()
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
