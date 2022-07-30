package listener

import (
	"testing"

	"github.com/altid/libs/service/store"
)

func TestListener(t *testing.T) {
	// Initiate a listener
	// Negotiate auth
	// Have a client connect
	// Do things like read and write files?
	np := Listen9p{}
	store := store.NewRamStorage()
	np.Register(store)
}