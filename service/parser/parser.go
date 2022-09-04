package parser

import (
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/internal/parse"
)

func ParseCtrlFile(b []byte) ([]*commander.Command, error) {
	return parse.ParseCtlFile(b)
}
