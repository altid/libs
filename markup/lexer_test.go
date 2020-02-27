package markup_test

import (
	"testing"

	"github.com/altid/libs/markup"
	fuzz "github.com/google/gofuzz"
)

func TestLexer(t *testing.T) {
	for i := 0; i < 50000; i++ {
		f := fuzz.New()

		var myData []byte
		f.Fuzz(&myData)
		l := markup.NewLexer(myData)

		for {
			i := l.Next()
			switch i.ItemType {
			case markup.EOF:
				return
			case markup.ErrorText:
				t.Error(i.Data)
			}
		}
	}
}
