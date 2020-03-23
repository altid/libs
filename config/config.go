package config

import (
	"os"
	"path"

	"github.com/altid/libs/fs"
	"github.com/mischief/ndb"
)

// Errors
const (
	ErrBadConfigurator = "configurator nil or invalid, cannot continue"
	ErrNoConfigure     = "unable to find or create config for this service. To create one, please run %s -conf"
	ErrNoSuchKey       = "no such key"
	ErrNoEntries       = "unable to find config entry for this service."
	ErrMultiEntries    = "config contains duplicate entries for this service. Please edit your altid/config file"
)

// Auth matches an auth= tuple in a config
// If the value matches factotum, it will use the factotum to return a password
// If the value matches password, it will return the value of a password= tuple
// If the value matches none, it will return an empty string
type Auth string

// Logdir is the directory that an Altid service can optionally store logs to
// If this is unset in the config, it will be filled with "none"
type Logdir string

// ListenAddress is the listen_address tuple in a config
// If this is unset in the config, it will be filled with "localhost"
type ListenAddress string

// Marshal will take a pointer to a struct as input, as well as the name of the service and attempt to fill the struct.
// - The struct entries must be of the type string, int, bool, or tls.Certificate
// - The tags of the struct will be used to indicate a query sent to the user
// - Default entries to the struct will be used as defaults
//
//	type myconf struct {
//		Name string `Username to use for the service`
// 		Port int `Port to connect with`
// 		UseSSL bool `Do you want to connect with SSL?`
// 		Auth config.Auth `Auth mechanism to use: password|factotum|none`
//		Logdir config.Logdir
//      Address config.ListenAddress
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
		debugLog = logger
	}

	// list all existing config entries
	have, err := fromConfig(debugLog, service, confdir)
	if err != nil {
		return err
	}

	if e := fillRequests(debugLog, requests, have); e != nil {
		return e
	}

	return nil
}

// Create takes a pointer to a struct, and attempts to create a config file entry on disk based on the struct
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
		debugLog = logger
	}

	have, err := fromConfig(debugLog, service, confdir)

	// Make sure we correct any errors we encounter
	switch {
	case os.IsNotExist(err):
		dir, err := fs.UserConfDir()
		if err != nil {
			return err
		}

		os.MkdirAll(path.Join(dir, "altid"), 0755)
		os.Create(getConf(service))
		debugLog("creating config file")

	// If we have multiple entries, something has indeed gone wrong
	// The user needs to manually clean this up
	case err.Error() == ErrMultiEntries:
		return err
		
	// This is the expected case in this situation
	case err.Error() == ErrNoEntries:
		debugLog("creating entry")
	}

	want, err := fromRequest(requests)
	if err != nil {
		return err
	}

	c, err := createConfFile(debugLog, service, have, want)
	if err != nil {
		return err
	}

	return c.writeToFile()
}

// GetLogDir returns a canonical directory for a user log, searching first altid/config
// If no entry is found or the file is missing, it will return "none"
func GetLogDir(service string) string {
	conf, err := ndb.Open(getConf(service))
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

	conf, err := ndb.Open(getConf(""))
	if err != nil {
		return nil, err
	}

	for _, rec := range conf.Search("service", "") {
		configs = append(configs, rec[0].Val)
	}

	return configs, nil
}
