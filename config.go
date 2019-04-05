package fslib

import (
	"log"
	"path"

	"github.com/mischief/ndb"
)

// GetLogDir returns a canonical directory for a user log, searching first altid/config
// If no entry is found or the file is missing, it will return a default path depending on the current operating system. Refer to UserShareDir documentation for what that is for your system
func GetLogDir(service string) string {
	confdir, err := UserConfDir()
	if err != nil {
		log.Fatal(err)
	}
	filePath := path.Join(confdir, "altid", "config")
	conf, err := ndb.Open(filePath)
	if err != nil {
		return logDir(service)
	}
	logdir := conf.Search("service", service).Search("log")
	if logdir != "" {
		return path.Join(logdir, service)
	}
	return logDir(service)
}

func logDir(service string) string {
	userdir, err := UserShareDir()
	if err != nil {
		log.Fatal(err)
	}
	return path.Join(userdir, "altid", service)
}
