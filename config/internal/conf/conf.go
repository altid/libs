package conf

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
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

func WriteOut(service string, request any) error {
	m := make(map[string]any)
	m["service"] = service
	m["altid_config_path"] = util.GetConf(service)

	switch runtime.GOOS {
	case "plan9":
		m["factotum_setup"] = "Set auth to factotum to avoid plaintext passwords"
		m["listen_address_link"] = "https://altid.github.io/using-listen-address.html#plan9"
	default:
		m["factotum_setup"] = "To set up factotum using plan9port to avoid plaintext passwords\n\t# see https://9fans.github.io/plan9port/man/man4/factotum.html"
		m["listen_address_link"] = "https://altid.github.io/using-listen-address.html"
	}

	// Walk our marshalled object and fill out our form
	s := reflect.ValueOf(request)
	t := reflect.Indirect(s).Type()
	for i := 0; i < t.NumField(); i++ {
		k := t.FieldByIndex([]int{i})
		d := reflect.Indirect(s).Field(i)
		//if k.Name == "auth" && d.CanSet() {
		//	// Set the auth=password if we have a password
		//}
		switch d.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			m[strings.ToLower(k.Name)] = fmt.Sprintf("%d", d.Int())
		case reflect.String:
			m[strings.ToLower(k.Name)] = d.String()
		}
	}

	tp := template.Must(template.New("entry").Parse(tmpl))

	return tp.Execute(os.Stdout, m)
}
