package config

import (
	"crypto/tls"
	"errors"
	"log"
	"os"
	"path"

	"github.com/altid/libs/auth"
	"github.com/altid/libs/fs"
	"github.com/mischief/ndb"
)

func getConf(service string) string {
	confdir, err := fs.UserConfDir()
	if err != nil {
		log.Fatal(err)
	}

	return path.Join(confdir, "altid", "config")
}

func findAuth(debug func(string, ...interface{}), service string, c ndb.RecordSet) (string, error) {
	switch c.Search("auth") {
	case "password":
		pass := c.Search("password")
		if pass == "" {
			return "", errors.New("unable to find password")
		}

		return pass, nil
	case "factotum":
		debug("request key=\"factotum\"")
		userPwd, err := auth.Getuserpasswd("proto=pass service=%s", service)
		if err != nil {
			debug("response key=\"password\" error=\"%v\"", err)
			return "", err
		}

		debug("response key=\"factotum\" value=\"success\"")
		return userPwd.Password, nil
	// If we're here, either we have a "none" value, or auth isn't listed
	default:
		return "", nil
	}
}

func findTlsCert(debug func(string, ...interface{}), c ndb.RecordSet) (tls.Certificate, error) {
	debug("request type=tls")
	cert := c.Search("cert")
	key := c.Search("key")

	if cert == "" || key == "" {
		return tls.Certificate{}, errors.New("missing cert/key entries in config")
	}
	return tls.LoadX509KeyPair(cert, key)
}

func findLogdir(debug func(string, ...interface{}), c ndb.RecordSet) string {
	debug("request key=\"log\"")

	dir := c.Search("log")
	if dir == "" {
		debug("response key=\"log\" value=\"none\"")
		return "none"
	}

	debug("response key=\"log\" value=\"%s\"", dir)
	return dir
}

func findListen(debug func(string, ...interface{}), c ndb.RecordSet) string {
	debug("request key=\"listen_address\"")
	dir := c.Search("listen_address")
	if dir == "" {
		debug("response key=\"listen_address\" value=\"none\"")
		return "none"
	}

	debug("response key=\"listen_address\" value=\"%s\"", dir)
	return dir
}

func logger(format string, v ...interface{}) {
	l := log.New(os.Stdout, "config: ", 0)
	l.Printf(format+"\n", v...)
}
