package html

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

type testHandler struct {
	c *Cleaner
}

func (t *testHandler) Nav(url *URL) error {
	_, err := fmt.Fprintf(t.c, "%s", url)
	return err
}

func (t *testHandler) Img(img *Image) error {
	_, err := fmt.Fprintf(t.c, "%s", img)
	return err
}

func TestNav(t *testing.T) {
	testParse(t, "_nav")
}

func TestImg(t *testing.T) {
	testParse(t, "_img")
}

func TestNormal(t *testing.T) {
	testParse(t, "")
}

/* TODO: test all markup elements
func TestElements(t *testing.T) {
	testParse(t, "_elements")
}
*/

func testParse(t *testing.T, target string) {
	p := &testHandler{}

	rp, err := os.Open(fmt.Sprintf("resources/test%v.html", target))
	if err != nil {
		t.Errorf("fatal system error: %v", err)
		t.Log("please ensure your local repository contains a resources/ folder")
		return
	}

	defer rp.Close()

	wp, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Errorf("fatal system error: %v", err)
		return
	}

	defer os.Remove(wp.Name())

	p.c, err = NewCleaner(wp, p)
	if err != nil {
		t.Errorf("library error: %v", err)
		return
	}

	defer p.c.Close()

	if e := p.c.Parse(rp); e != nil && e != io.EOF {
		t.Errorf("parsing error: %v", err)
		return
	}

	result, err := ioutil.ReadFile(wp.Name())
	if err != nil {
		t.Errorf("system error: %v", err)
		return
	}

	expected, err := ioutil.ReadFile(fmt.Sprintf("resources/test%s.md", target))
	if err != nil {
		t.Errorf("system error: %v", err)
	}

	if !(bytes.Equal(result, expected)) {
		t.Error("parsing failed, bytes did not match expected output")
	}
}
