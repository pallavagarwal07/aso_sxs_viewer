package createwindow

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestDisableCrashedBubble(t *testing.T) {
	content := []byte("\"exit_type\":\"Crashed\"")

	tmpfile, err := ioutil.TempFile("", "tempfile")
	if err != nil {
		t.Errorf("Encountered error %s in creating temp directory", err)
	}

	if _, err = tmpfile.Write(content); err != nil {
		t.Errorf("Encountered error %s in writing content to temp file", err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Errorf("Encountered error %s in closing temp file", err)
	}

	if err := DisableCrashedBubble(tmpfile.Name()); err != nil {
		t.Errorf("Encountered error %s in DisableCrashedBubble", err)
	}
	os.Remove(tmpfile.Name())
}
