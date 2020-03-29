package markup_test

import (
	"fmt"
	"testing"

	"github.com/altid/libs/markup"
	fuzz "github.com/google/gofuzz"
)

func TestSpecial(t *testing.T) {
	myData := []byte("A string with an un-escaped hyphen")
	l := markup.NewLexer(myData)

	if _, e := l.String(); e == nil {
		t.Error("no error on incorrectly formatted text")
	}

	myData = []byte("A string with an un\\-un\\-escaped hypthen")
	l = markup.NewLexer(myData)

	if _, e := l.String(); e != nil {
		t.Error(e)
	}

	myData = []byte("A %[Test in a sub-tag](red) should fail")
	l = markup.NewLexer(myData)

	s, e := l.String()
	if e == nil {
		t.Error("no error on incorrectly formatted text")
	}

	fmt.Println(s)
}

func TestLexer(t *testing.T) {
	for i := 0; i < 50000; i++ {
		f := fuzz.New()

		myData := make([]byte, 50)
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
