package createwindow

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestDisableCrashedBubble(t *testing.T) {
	content, err := ioutil.ReadFile("testfiles/testdata.txt")

	if err != nil {
		t.Errorf("Encountered error %s in reading test file", err)
	}

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
		{
			t.Errorf("Encountered error %s in DisableCrashedBubble", err)
		}
	}

	newtmpfile, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Errorf("Encountered error %s in reading test file", err)
	}

	isReplaced := strings.Contains(string(newtmpfile), "\"exit_type\":\"Normal\"") && !strings.Contains(string(newtmpfile), "\"exit_type\":\"Crashed\"")
	if !isReplaced {
		t.Errorf("Encountered error in replacing string")
	}
	os.Remove(tmpfile.Name())
}
