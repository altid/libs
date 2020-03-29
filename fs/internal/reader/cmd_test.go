package reader

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestCmd(t *testing.T) {
	// Make a ctl file
	// Read from it
	// Send in goroutine
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error(err)
	}

	rd, err := Cmd(dir)
	if err != nil {
		t.Error(err)
	}

	fp, err := os.OpenFile(path.Join(dir, "ctl"), os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Error(err)
		return
	}

	go sendData(fp)

	lines := bufio.NewReader(rd)
	for i := 0; i < 100; i++ {
		if _, e := lines.ReadString('\n'); e != nil {
			t.Error(e)
		}
	}
}
