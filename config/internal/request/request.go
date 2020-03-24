package request

import (
	"fmt"
	"reflect"

	"github.com/altid/libs/config/types"
)

type Request struct {
	Key      string
	Prompt   string
	Defaults interface{}
}

func Build(req interface{}) ([]*Request, error) {
	var reqs []*Request

	s := reflect.ValueOf(req)
	t := reflect.Indirect(s).Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.FieldByIndex([]int{i})
		req := &Request{
			Key:    f.Name,
			Prompt: string(f.Tag),
		}

		d := reflect.Indirect(s).Field(i)
		switch f.Type.Name() {
		case "Auth":
			req.Defaults = d.Interface().(types.Auth)
		case "Logdir":
			req.Defaults = d.Interface().(types.Logdir)
		case "ListenAddress":
			req.Defaults = d.Interface().(types.ListenAddress)
		case "string":
			req.Defaults = d.String()
		case "int":
			req.Defaults = d.Int()
		case "bool":
			req.Defaults = d.Bool()
		default:
			return nil, fmt.Errorf("unknown type for config: %s", f.Type.Name())
		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}
