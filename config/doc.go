/*Package config allows Altid services to interact withe the ndb-formatted configuration files used by Altid

	go get github.com/altid/libs/config

The ndb format is described in http://man.cat-v.org/plan_9/6/ndb

Usage

A service can marshall values to a struct through Marshall, as long as the entries follow these rules:
- Nested structs will be ignored
- Type must be bool, int, string, or one of Auth, Log, or ListenAddress
- structs must have default values set

Example:

	// package mypackage

	import (
		"flag"
		"log"
		"os"
		"os/user"

		"github.com/altid/libs/config"
		"github.com/altid/libs/config/types"
	)

	var conf = flag.Bool("conf", false, "Create configuration file")

	func main() {
		flag.Parse()

		u, _ := user.Current

		mytype := struct {
			// Struct tags are used by Create to interactively fill in any missing data
			Name string `Enter a name to use on the service`
			UseTLS bool `Use TLS? (true|false)`
			Port int
			Auth types.Auth `Enter the authentication method you would like to use: password|factotum|none`
			Logdir types.Logdir
			ListenAddress config.ListenAddress
		}{u.Name, false, 564, "none", "", ""}

		if flag.Lookup("conf") != nil {
			if e := config.Marshall(&mytype, "myservice", false); e != nil {
				log.Fatal(e)
			}

			os.Exit(0)
		}

		// Your error message should indicate that the user re-runs with -conf to create any missing entries
		if e := config.Marshall(&mytype, "myservice", false); e != nil {
			log.Fatal("unable to create config: %v\nRun program with -conf to create missing entries")
		}

		// [...]
	}

*/
package config
