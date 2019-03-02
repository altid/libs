package fslib

import (
	"log"
	"path"

	"github.com/mischief/ndb"
)

func GetLogDir(service string) string {
	confdir, err := UserConfDir()
	if err != nil {
		log.Fatal(err)
	}
	filePath := path.Join(confdir, "ubqt.cfg")
	conf, err := ndb.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	logdir := conf.Search("service", service).Search("log")
	if logdir != "" {
		return path.Join(logdir, service)
	}
	userdir, err := UserShareDir()
	if err != nil {
		log.Fatal(err)
	}
	return path.Join(userdir, "ubqt", service)
}
