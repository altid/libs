package service_test

import (
	"github.com/altid/libs/service"
	"github.com/altid/libs/service/listener"
)

func ExampleNew() {
	l := listener.Listen9p{}
	s := &service.Command{}
	service.New(s, l, "", false)
}
