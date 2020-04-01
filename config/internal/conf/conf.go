package conf

import (
	"os"
	"runtime"
	"text/template"

	"github.com/altid/libs/config/internal/entry"
	"github.com/altid/libs/config/internal/util"
	"github.com/altid/libs/config/types"
)

const tmpl = `Configuration successful!
Please add the following line(s) to your {{.altid_config_path}}

# Ensure only one service={{.service}} line is present in the file
service={{.service}}{{range $key, $value := .}}{{if eq $key "service"}}{{else if eq $key "altid_config_path"}}{{else if eq $key "factotum_setup"}}{{else if eq $key "listen_address_link"}}{{else if eq $key "port"}}{{else if eq $key "password"}}{{else if eq $key "listen"}}{{else if eq $key "log"}}{{else if eq $key "tlscert"}}{{else if eq $key "listenaddress"}}{{else if eq $key "tlskey"}}{{else}} {{$key}}={{$value}}{{end}}{{end}}
{{if .listen}}	# More info {{.listen_address_link}}
	listen_address={{.listen}}
{{end}}{{if .listenaddress}}	# More info {{.listen_address_link}}
	listen_address={{.listenaddress}}
{{end}}{{if .port}}	port={{.port}}
{{end}}{{if .log}}	log={{.log}}
{{end}}{{if .password}}	# {{.factotum_setup}}
	password={{.password}}
{{end}}{{if .tlskey}}	tlskey={{.tlskey}}
{{end}}{{if .tlscert}}	tlscert={{.tlscert}}
{{end}}`

type Conf struct {
	name    string
	entries []*entry.Entry
}

func FromConfig(debug func(string, ...interface{}), service string, confdir string) ([]*entry.Entry, error) {
	return entry.FromConfig(debug, service, confdir)
}

// FixAuth is a helper func to correct the auth being set to the value of password in FromConfig
func FixAuth(have []*entry.Entry, service, cfgfile string) {
	for _, ent := range have {
		if ent.Key != "auth" {
			continue
		}

		value := entry.FixAuth(service, cfgfile)

		// ndb lookup returns string
		ent.Value = types.Auth(value.Value.(string))
	}
}

func (c *Conf) WriteOut() error {
	m := make(map[string]interface{})
	m["service"] = c.name
	m["altid_config_path"] = util.GetConf(c.name)

	switch runtime.GOOS {
	case "plan9":
		m["factotum_setup"] = "Set auth to factotum to avoid plaintext passwords"
		m["listen_address_link"] = "https://altid.github.io/using-listen-address.html#plan9"
	default:
		m["factotum_setup"] = "To set up factotum using plan9port to avoid plaintext passwords\n\t# see https://9fans.github.io/plan9port/man/man4/factotum.html"
		m["listen_address_link"] = "https://altid.github.io/using-listen-address.html"
	}

	for _, entry := range c.entries {
		m[entry.Key] = entry.String()
	}

	tp := template.Must(template.New("entry").Parse(tmpl))

	return tp.Execute(os.Stdout, m)
}
