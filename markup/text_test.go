package markup_test

import (
	"fmt"
	"testing"

	"github.com/altid/libs/markup"
)

var path = []byte("https://github.com")
var alt = []byte("a photo of octocat")
var id = []byte("host for code online")
var none = []byte("")

func TestColor(t *testing.T) {
	c, err := markup.NewColor("#FFF", []byte("test"))
	if err != nil || c.String() != "%[test](#FFF)" {
		t.Error("failed to parse hex colour code")
	}

	c, err = markup.NewColor("#FFFFFF", []byte("test"))
	if err != nil || c.String() != "%[test](#FFFFFF)" {
		t.Error("failed to parse hex colour code")
	}

	c, err = markup.NewColor("grey", []byte("test"))
	if err != nil || c.String() != "%[test](grey)" {
		t.Error("failed to parse colour codes")
	}

	_, err = markup.NewColor("chicken", []byte("test"))
	if err == nil {
		t.Error("parsing error - invalid code in NewColor")
	}
}

func TestUrl(t *testing.T) {
	c, err := markup.NewURL(path, id)
	if err != nil || c.String() != fmt.Sprintf("[%s](%s)", id, path) {
		t.Error("failed to parse URL")
	}

	c, err = markup.NewURL(path, none)
	if err != nil || c.String() != fmt.Sprintf("[%s](%s)", path, path) {
		t.Error("failed to parse URL")
	}

	_, err = markup.NewURL(none, none)
	if err == nil {
		t.Error("parsing error - invalid code in NewUrl")
	}
}

func TestNotifier(t *testing.T) {
	c := markup.NewNotifier("#foo", "bar - the place to be!", "baz is a *real* friend")
	foo, bar, baz := c.Parse()

	if foo != "foo" && bar != "# bar \\- the place to be\\!" && baz != "baz is a \\*real\\* friend" {
		t.Error("parsing error - invalid code in Notifier.Parse")
	}
}

func TestEscapeString(t *testing.T) {
	c := "this *is* ~my _test_ string~ to -see-"
	if markup.EscapeString(c) != "this \\*is\\* \\~my \\_test\\_ string\\~ to \\-see\\-" {
		t.Error("parsing error in EscapeString")
	}
}
