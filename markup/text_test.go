package markup

import (
	"fmt"
	"testing"
)

var path = []byte("https://github.com")
var alt = []byte("a photo of octocat")
var id = []byte("host for code online")
var none = []byte("")

func TestColor(t *testing.T) {
	c, err := NewColor("#FFF", []byte("test"))
	if err != nil || c.String() != "%[test](#FFF)" {
		t.Error("failed to parse hex colour code")
	}

	c, err = NewColor("#FFFFFF", []byte("test"))
	if err != nil || c.String() != "%[test](#FFFFFF)" {
		t.Error("failed to parse hex colour code")
	}

	c, err = NewColor("grey", []byte("test"))
	if err != nil || c.String() != "%[test](grey)" {
		t.Error("failed to parse colour codes")
	}

	_, err = NewColor("chicken", []byte("test"))
	if err == nil {
		t.Error("parsing error - invalid code in NewColor")
	}
}

func TestUrl(t *testing.T) {
	c, err := NewURL(path, id)
	if err != nil || c.String() != fmt.Sprintf("[%s](%s)", id, path) {
		t.Error("failed to parse URL")
	}

	c, err = NewURL(path, none)
	if err != nil || c.String() != fmt.Sprintf("[%s](%s)", path, path) {
		t.Error("failed to parse URL")
	}

	_, err = NewURL(none, none)
	if err == nil {
		t.Error("parsing error - invalid code in NewUrl")
	}
}

func TestNotifier(t *testing.T) {
	c := NewNotifier("#foo", "bar - the place to be!", "baz is a *real* friend")
	foo, bar, baz := c.Parse()

	if foo != "foo" && bar != "# bar \\- the place to be\\!" && baz != "baz is a \\*real\\* friend" {
		t.Error("parsing error - invalid code in Notifier.Parse")
	}
}

func TestEscapeString(t *testing.T) {
	c := "this *is* ~my _test_ string~ to -see-"
	if EscapeString(c) != "this \\*is\\* \\~my \\_test\\_ string\\~ to \\-see\\-" {
		t.Error("parsing error in EscapeString")
	}
}
