package html

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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

	wp, err := os.CreateTemp("", "test")
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

	result, err := os.ReadFile(wp.Name())
	if err != nil {
		t.Errorf("system error: %v", err)
		return
	}

	expected, err := os.ReadFile(fmt.Sprintf("resources/test%s.md", target))
	if err != nil {
		t.Errorf("system error: %v", err)
	}

	if !(bytes.Equal(result, expected)) {
		t.Error("parsing failed, bytes did not match expected output")
	}
}

// From https://github.com/adtile/fixed-nav/blob/master/index.html (MIT)
func TestParseNav(t *testing.T) {
	input := `<!DOCTYPE html>
	<html lang="en">
	  <head>
		<meta charset="utf-8">
		<title>Adtile Fixed Nav</title>
		<link rel="stylesheet" href="css/styles.css">
	  </head>
	  <body>
	
		<header>
		  <a href="#home" class="logo" data-scroll>Nav test</a>
		  <nav class="nav-collapse">
			<ul>
			  <li class="menu-item active"><a href="#home" data-scroll>Home</a></li>
			  <li class="menu-item"><a href="#about" data-scroll>About</a></li>
			  <li class="menu-item"><a href="#projects" data-scroll>Projects</a></li>
			  <li class="menu-item"><a href="#blog" data-scroll>Blog</a></li>
			  <li class="menu-item"><a href="http://www.google.com" target="_blank">Google</a></li>
			</ul>
		  </nav>
		</header>
	
		<section id="home">
		  <h1>Fixed Nav</h1>
		  <p>The code and examples are hosted on GitHub and can be <a href="https://github.com/adtile/fixed-nav">found from here</a>. Read more about the approach from&nbsp;<a href="http://blog.adtile.me/2014/03/03/responsive-fixed-one-page-navigation/">our&nbsp;blog</a>.</p>
		</section>
	
		<section id="about">
		  <h1>About</h1>
		</section>
	
		<section id="projects">
		  <h1>Projects</h1>
		</section>
	
		<section id="blog">
		  <h1>Blog</h1>
		</section>
	
		<script src="js/fastclick.js"></script>
		<script src="js/scroll.js"></script>
		<script src="js/fixed-responsive-nav.js"></script>
	  </body>
	</html>`

	//c, _ := NewCleaner(os.Stdout, &testHandler{})
	z := html.NewTokenizer(strings.NewReader(input))
	for {
		z.Next()

		if z.Err() != nil {
			break
		}

		if z.Token().DataAtom != atom.Nav {
			continue
		}

		i := 0

		for nav := range parseNav(z) {
			if len(nav.Href) < 1 || len(nav.Msg) < 1 {
				t.Errorf("unable to parse element %s", nav.String())
			}
			i++
		}

		if i != 5 {
			t.Error("failed to correctly parse nav")
			return
		}

		return
	}

	t.Error("Did not find nav element (upstream error)")
}
