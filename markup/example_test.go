package markup_test

import (
	"github.com/altid/libs/markup"
)

func ExampleLexer() {
	l := markup.NewStringLexer("Text with **bold and _strong tags_**")
	for {
		n := l.Next()
		switch n.ItemType {
		case markup.URLText:
		case markup.BoldText:
		case markup.StrongText:
		case markup.EOF:
			break
		}
	}
}
