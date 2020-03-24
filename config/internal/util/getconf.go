package util

import (
	"log"
	"path"

	"github.com/altid/libs/fs"
)

func GetConf(service string) string {
	confdir, err := fs.UserConfDir()
	if err != nil {
		log.Fatal(err)
	}

	return path.Join(confdir, "altid", "config")
}
