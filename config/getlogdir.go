package config

import (
	"path"

	"github.com/mischief/ndb"
)

// GetLogDir returns a canonical directory for a user log, searching first altid/config
// If no entry is found or the file is missing, it will return "none"
func GetLogDir(service string) string {
	conf, err := ndb.Open(getConfDir(service))
	if err != nil {
		return "none"
	}

	logdir := conf.Search("service", service).Search("log")
	if logdir != "" {
		return path.Join(logdir, service)
	}

	return "none"
}
