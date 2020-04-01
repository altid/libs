package build

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/altid/libs/config/internal/entry"
	"github.com/altid/libs/config/internal/request"
	"github.com/altid/libs/config/types"
)

func Marshal(debug func(string, ...interface{}), requests interface{}, have []*entry.Entry) error {
	// Loop through entries and attempt to fill the struct
	// Make sure we turn any ints that are supposed to be strings,
	// back into strings!
	want, err := request.Build(requests)
	if err != nil {
		return err
	}

	for _, item := range want {
		item.Key = strings.ToLower(item.Key)

		if entry, ok := entry.Find(item, have); ok {
			if e := push(requests, entry); e != nil {
				return e
			}

			continue
		}

		if item.Defaults != nil {
			entry := &entry.Entry{
				Key:   item.Key,
				Value: item.Defaults,
			}

			if e := push(requests, entry); e != nil {
				return e
			}

			continue
		}

		return fmt.Errorf("missing config entry for %s", item.Key)
	}
	return nil
}

// Add the actual value to the struct, being careful that
// the strconv.Atoi from before is guarded against
func push(requests interface{}, entry *entry.Entry) error {
	s := reflect.ValueOf(requests)
	t := reflect.Indirect(s).Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.FieldByIndex([]int{i})

		tag := f.Tag.Get("altid")
		if tag == "" {
			continue
		}

		// Use the tag, Luke
		req, err := request.ParseTag(tag)
		if err != nil {
			continue
		}

		if req.Key != entry.Key {
			continue
		}

		d := reflect.Indirect(s).Field(i)
		switch f.Type.Name() {
		case "string":
			d.SetString(entry.Value.(string))
		case "Auth":
			d.Set(reflect.ValueOf(entry.Value.(types.Auth)))
		case "Logdir":
			d.Set(reflect.ValueOf(entry.Value.(types.Logdir)))
		case "ListenAddress":
			d.Set(reflect.ValueOf(entry.Value.(types.ListenAddress)))
		default:
			d.Set(reflect.ValueOf(entry.Value))
		}

		return nil
	}

	return fmt.Errorf("could not add %s entry to struct", entry.Key)
}
