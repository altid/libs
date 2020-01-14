package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
)

// ctl files need to parse commands before sending them on
type control struct {
	service string
	buff    string
	cmd     string
	payload string
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
