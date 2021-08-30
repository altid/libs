package reader

import (
	"bufio"
	"crypto/rand"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"
)

func sendData(f *os.File) {
	for i := 0; i < 100; i++ {
		f.WriteString("line\n")
		n, _ := rand.Int(rand.Reader, big.NewInt(15))
		time.Sleep(time.Duration(n.Int64()) * time.Millisecond)
	}

	f.Close()
}

func TestPoller(t *testing.T) {
	// Make a ctl file
	// Read from it
	// Send in goroutine
	file, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}

	rd, err := Poll(file.Name())
	if err != nil {
		t.Error(err)
	}

	go sendData(file)
	defer rd.Close()
	b := bufio.NewReader(rd)

	for i := 0; i < 100; i++ {
		_, err := b.ReadString('\n')
		if err != nil {
			t.Error(err)
			break
		}
	}
}
