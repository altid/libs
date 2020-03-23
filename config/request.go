package config

import (
	"crypto/tls"
	"fmt"
	"reflect"
	"strings"
)

type request struct {
	key      string
	prompt   string
	defaults interface{}
}

func fromRequest(req interface{}) ([]*request, error) {
	var reqs []*request

	s := reflect.ValueOf(req)
	t := reflect.Indirect(s).Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.FieldByIndex([]int{i})
		req := &request{
			key:    f.Name,
			prompt: string(f.Tag),
		}

		d := reflect.Indirect(s).Field(i)
		switch f.Type.Name() {
		case "Auth":
			req.defaults = d.Interface().(Auth)
		case "Logdir":
			req.defaults = d.Interface().(Logdir)
		case "ListenAddress":
			req.defaults = d.Interface().(ListenAddress)
		case "Certificate":
			req.defaults = d.Interface().(tls.Certificate)
		case "string":
			req.defaults = d.String()
		case "int":
			req.defaults = d.Int()
		case "bool":
			req.defaults = d.Bool()
		default:
			return nil, fmt.Errorf("unknown type for config: %s", f.Type.Name())
		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}

func fillRequests(debug func(string, ...interface{}), requests interface{}, have []*entry) error {
	// Loop through entries and attempt to fill the struct
	// Make sure we turn any ints that are supposed to be strings,
	// back into strings!
	want, err := fromRequest(requests)
	if err != nil {
		return err
	}

	for _, item := range want {
		item.key = strings.ToLower(item.key)

		if entry, ok := haveEntry(debug, item, have); ok {
			if e := push(requests, entry); e != nil {
				return e
			}

			continue
		}

		return fmt.Errorf("missing config entry for %s", item.key)
	}
	return nil
}

// Add the actual value to the struct, being careful that
// the strconv.Atoi from before is guarded against
func push(requests interface{}, entry *entry) error {
	s := reflect.ValueOf(requests)
	t := reflect.Indirect(s).Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.FieldByIndex([]int{i})

		if strings.ToLower(f.Name) != entry.key {
			continue
		}

		d := reflect.Indirect(s).Field(i)
		switch f.Type.Name() {
		case "string":
			switch v := entry.value.(type) {
			// Make sure we cast back to string when
			// the struct wanted a numerical string
			case int:
				d.SetString(fmt.Sprintf("%d", v))
			case string:
				d.SetString(v)
			default:
				d.Set(reflect.ValueOf(entry.value))
			}
		case "Auth":
			d.Set(reflect.ValueOf(Auth(entry.value.(string))))
		case "Logdir":
			d.Set(reflect.ValueOf(Logdir(entry.value.(string))))
		case "ListenAddress":
			d.Set(reflect.ValueOf(ListenAddress(entry.value.(string))))
		case "Certificate":
			d.Set(reflect.ValueOf(entry.value.(tls.Certificate)))
		default:
			d.Set(reflect.ValueOf(entry.value))
		}

		return nil
	}

	return fmt.Errorf("could not add %s entry to struct", entry.key)
}
