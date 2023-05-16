package store_test

import (
	"log"
	"os"

	"github.com/altid/libs/store"
)

func ExampleNewRamstore() {
	rs := store.NewRamstore(false)
	f, err := rs.Open("testfile")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.Write([]byte("Some data"))
}

func ExampleNewLogstore() {
	tmp, err := os.MkdirTemp("", "altid")
	if err != nil {
		log.Fatal(err)
	}
	rs := store.NewLogstore(tmp, false)
	f, err := rs.Open("test/main")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.Write([]byte("Some data"))
}
