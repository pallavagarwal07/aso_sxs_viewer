package command

import (
	"fmt"
	"os"
	"testing"
)

func TestExecuteProgram(t *testing.T) {

	test := ExternalCommand{
		Path: "./test_command",
		Env:  os.Environ(),
	}

	programstate, err := ExecuteProgram(test, testErrorHandler)
	if err != nil {
		t.Errorf("Encountered error %s by ExecuteProgram()", err.Error())
	}
	out := programstate.Stdout()

	serr := programstate.Stderr()
	if len(serr) != 0 {
		fmt.Println("Returned error by StderrNonBlocking()", string(serr))
	}

	want := "Hello"
	got := string(out)

	if got != want {
		t.Errorf("Got text = %q, want %q", got, want)
	}

	err = programstate.Command.Process.Signal(os.Interrupt)
	if err != nil {
		t.Errorf("Encountered error %s while sending interrupt signal", err.Error())
	}

	out = programstate.Stdout()

	want = "Hello\nWorld"
	got = string(out)

	if got != want {
		t.Errorf("Got text = %q, want %q", got, want)
	}
}

func testErrorHandler(p *ProgramState, err error) error {
	fmt.Println("Test_error_handler", err)
	return err
}
