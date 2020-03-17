/*Package config allows Altid services to interact withe the ndb-formatted configuration files used by Altid

	go get github.com/altid/libs/config

The ndb format is described in http://man.cat-v.org/plan_9/6/ndb

Usage

This library has some special usage notes. Namely, on error it will return a string indicating the user should run `myservice -conf`. 
Doing so will call the Configurator which was passed in to config.New() and exit the program. 

Repl

The included Repl is meant to make designing Configurators much easier. To use, simply pass a struct to Repl with all the entries you wish described as string entries
The struct tags will be used verbatim as the prompt message to the client

	func myConfigurator(rw io.ReadCloser) (*config.Config, error) {
		repl := struct {
			Address string `IP Address of service`
			Port int `Port to use`
			UseTLS string `Do you wish to use TLS? yes/no`
		}{"localhost", 564, "false"}

		return config.Repl(rw, repl, false)
	}

The library will walk through each step, in order, and build a config entry.
*/
package config