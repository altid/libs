package store_test

import (
	"log"
	"os"

	"github.com/altid/libs/store"
)

func ExampleNewRamStore() {
	rs := store.NewRamStore()
	f, err := rs.Open("testfile")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.Write([]byte("Some data"))
}

func ExampleNewLogStore() {
	tmp, err := os.MkdirTemp("", "altid")
	if err != nil {
		log.Fatal(err)
	}
	rs := store.NewLogStore(tmp)
	f, err := rs.Open("test/main")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.Write([]byte("Some data"))
}
