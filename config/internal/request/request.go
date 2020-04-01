package request

import (
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/altid/libs/config/types"
)

type Request struct {
	Field    string
	Key      string
	Prompt   string
	Defaults interface{}
	Pick     []string
}

func Build(req interface{}) ([]*Request, error) {
	var reqs []*Request

	s := reflect.ValueOf(req)
	t := reflect.Indirect(s).Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.FieldByIndex([]int{i})

		// Not in an altid tag, skip
		tag := f.Tag.Get("altid")
		if tag == "" {
			continue
		}

		req, err := ParseTag(tag)
		if err != nil {
			return nil, err
		}

		req.Field = f.Name
		d := reflect.Indirect(s).Field(i)

		switch f.Type.Name() {
		case "string":
			req.Defaults = d.String()
		case "Auth":
			req.Defaults = d.Interface().(types.Auth)
		case "Logdir":
			req.Defaults = d.Interface().(types.Logdir)
		case "ListenAddress":
			req.Defaults = d.Interface().(types.ListenAddress)
		case "int":
			req.Defaults = d.Interface().(int)
		case "uint":
			req.Defaults = d.Interface().(uint)
		case "int8":
			req.Defaults = d.Interface().(int8)
		case "uint8":
			req.Defaults = d.Interface().(uint8)
		case "int16":
			req.Defaults = d.Interface().(int16)
		case "uint16":
			req.Defaults = d.Interface().(uint16)
		case "int32":
			req.Defaults = d.Interface().(int32)
		case "uint32":
			req.Defaults = d.Interface().(uint32)
		case "int64":
			req.Defaults = d.Interface().(int64)
		case "uint64":
			req.Defaults = d.Interface().(uint64)
		case "float32":
			req.Defaults = d.Interface().(float32)
		case "float64":
			req.Defaults = d.Interface().(float64)
		case "bool":
			req.Defaults = d.Bool()
		default:
			return nil, fmt.Errorf("unknown type for config: %s", f.Type.Name())
		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}

func ParseTag(tag string) (*Request, error) {
	r := &Request{}

	// key prompt pick
	vals, err := csv.NewReader(strings.NewReader(tag)).Read()
	if err != nil {
		return nil, err
	}

	switch len(vals) {
	case 0:
		return nil, errors.New("tag contained no values")
	case 1:
		r.Key = vals[0]
		return r, nil
	}

	for _, val := range vals[1:] {
		switch {
		case val == "no_prompt":
			break
		case strings.HasPrefix(val, "prompt:"):
			r.Prompt = strings.TrimPrefix(val, "prompt:")
		case strings.HasPrefix(val, "pick:"):
			opts := strings.TrimPrefix(val, "pick:")
			r.Pick = strings.Split(opts, "|")
		}
	}

	r.Key = vals[0]
	return r, nil
}
