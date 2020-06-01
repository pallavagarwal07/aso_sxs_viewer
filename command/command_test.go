package command

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestExecuteProgram(t *testing.T) {

	cat := ExternalCommand{
		Path: "cat",
		Arg: []string{
			"test_sample.txt",
		},
		Env: os.Environ(),
	}

	programstate, err := ExecuteProgram(cat, testErrorHandler)
	if err != nil {
		t.Errorf("Encountered error %s by ExecuteProgram()", err.Error())
	}
	out, err := programstate.StdoutNonBlocking()
	if err != nil {
		t.Errorf("Encountered error %s by StdoutNonBlocking()", err.Error())
	}
	_, err = programstate.StderrNonBlocking()
	if err != nil {
		t.Errorf("Encountered error %s by StderrNonBlocking()", err.Error())
	}

	want := "Hello World"
	got := string(out)

	if got != want {
		t.Errorf("Got text = %q, want %q", got, want)
	}
}

func testErrorHandler(cmd *exec.Cmd, err error) error {
	fmt.Println("Test_error-handler")
	return err
}
