package markup

import (
	"testing"

	fuzz "github.com/google/gofuzz"
)

func TestSpecial(t *testing.T) {
	myData := []byte("A string with an un_escaped underscore")

	l := NewLexer(myData)

	if _, e := l.String(); e == nil {
		t.Error("no error on incorrectly formatted text")
	}

	myData = []byte("A string with an un-un-escaped hypthen")
	l = NewLexer(myData)

	if _, e := l.String(); e != nil {
		t.Error(e)
	}

	myData = []byte("Test **bold text** only")
	l = NewLexer(myData)

	if _, e := l.String(); e != nil {
		t.Error(e)
	}

	myData = []byte("Test **bold text* only")
	l = NewLexer(myData)

	if _, e := l.String(); e == nil {
		t.Error("no error on incorrectly formatted text")
	}

	myData = []byte("Test closing **bold and _strong tag_** correctly")
	l = NewLexer(myData)

	if _, e := l.String(); e != nil {
		t.Error(e)
	}

	myData = []byte("A %[Test in a sub_tag](red) should fail")
	l = NewLexer(myData)

	if _, e := l.String(); e == nil {
		t.Error("no error on incorrectly formatted text")
	}

	myData = []byte("A %[Test of **strong _text_ **](red)")
	l = NewLexer(myData)

	if _, e := l.String(); e != nil {
		t.Error(e)
	}

}

func TestLexer(t *testing.T) {
	for i := 0; i < 50000; i++ {
		f := fuzz.New()

		myData := make([]byte, 300)
		f.Fuzz(&myData)
		l := NewLexer(myData)

		for {
			i := l.Next()

			switch i.ItemType {
			case EOF:
				return
			// This is just interesting, really
			// Since we could theoretically get fuzzed data
			// That is incorrectly formatted
			// The real error is a timeout
			case ErrorText:
				t.Logf("Found some bad text from %s", myData)
				return
			}
		}
	}
}
