package store_test

import (
	"log"
	"os"
	"path"

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
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		log.Fatal(err)
	}
	rs := store.NewLogStore(tmp)
	f, err := rs.Open("test/main")
	defer f.Close()

	f.Write([]byte("Some data"))
	if _, err := os.ReadFile(path.Join(tmp, "test/main")); err != nil {
		log.Fatal("Error in reading datas")
	}
}
