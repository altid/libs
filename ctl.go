package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

// ctl files need to parse commands before sending them on
type control struct {
	service string
	buff    string
	cmd     string
	payload string
}

var ctl chan interface{}

func init() {
	in = make(chan interface{})
	s := &fileHandler{
		fn: newControl,
		ch: ctl,
	}
	addFileHandler("ctl", s)
}

func listenControls(ctx context.Context, cfg *config) (chan interface{}, error) {
	return ctl, nil
}

// Heavy lifting here with fields function and join should be rewritten eventually
func newControl(msg *message) interface{} {
	m := strings.Fields(msg.data)
	return &control{
		service: msg.service,
		cmd:     m[0],
		payload: strings.Join(m[1:], " "),
		buff:    msg.buff,
	}
}

func handleControl(ctx context.Context, msg interface{}) {
	ctl, ok := msg.(*control)
	if !ok {
		return
	}
	switch ctl.cmd {
	case "quit":
		ctx.Done()
	default:
		file := path.Join(*inpath, ctl.service, ctl.buff, "ctl")
		fp, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			return
		}
		defer fp.Close()
		fp.WriteString(fmt.Sprintf("%s %s", ctl.cmd, ctl.payload))
	}
}
