package command

import (
	"os"
	"testing"
)

const data = `general:
	open|close|banana       buffer buffer2  # Open and change buffers to a given service
	close   buffer  # Close a buffer and return to the last opened previously
	buffer  buffer  # Change to the named buffer
	link    to from # Overwrite the current <to> buffer with <from>, switching to from after. This destroys <to>
	quit    # Exits the service
media:
	play    track   # Play the named track
	pause   # Pause the current track
	next    # Play next

`

func TestParse(t *testing.T) {
	cmdlist, err := Parse([]byte(data))
	if err != nil {
		t.Error(err)
		return
	}

	fp, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0644)
	if e := PrintCtlFile(cmdlist, fp); e != nil {
		t.Error(err)
	}

}
