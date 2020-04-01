package entry

import (
	"log"
	"testing"
)

func TestFromConfig(t *testing.T) {
	// Make up a tmp file for testing
	db, err := FromConfig(log.Printf, "banana", "resources/config")
	if err != nil {
		t.Error(err)
	}

	if len(db) != 6 {
		t.Error("unable to find all entries")
	}
}
