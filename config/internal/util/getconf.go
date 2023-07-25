package util

import (
	"log"
	"path"

	"github.com/altid/libs/service"
)

func GetConf() string {
	confdir, err := service.UserConfDir()
	if err != nil {
		log.Fatal(err)
	}
	return path.Join(confdir, "altid", "config")
}
