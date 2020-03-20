package files

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

type mytype struct {
	path string
}

func TestFiles(t *testing.T) {
	fp, _ := ioutil.TempFile(os.TempDir(), "test")
	fp.WriteString("Bunch of test data")

	m := &mytype{fp.Name()}

	h := Handle(m.path)
	h.Add("/test", m)

	s, err := h.Stat("", "/test")
	if err != nil {
		t.Error(err)
		return
	}

	if s.Name() != path.Base(fp.Name()) {
		t.Error("names did not match")
	}

	d, err := h.Normal("", "/test")
	if err != nil {
		t.Error(err)
		return
	}

	df, ok := d.(*os.File)
	if !ok {
		t.Error("unable to convert to file")
	}

	df.Close()
}

func (m *mytype) Stat(msg *Message) (os.FileInfo, error) {
	return os.Lstat(m.path)
}

func (m *mytype) Normal(msg *Message) (interface{}, error) {
	return os.Open(m.path)
}
