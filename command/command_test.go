package command

import (
	"fmt"
	"os"
	"testing"
)

func TestExecuteProgram(t *testing.T) {

	test := ExternalCommand{
		Path: "./test_sample_program",
		Env:  os.Environ(),
	}

	programstate, err := ExecuteProgram(test, testErrorHandler)
	if err != nil {
		t.Errorf("Encountered error %s by ExecuteProgram()", err.Error())
	}
	out, err := programstate.StdoutNonBlocking()
	if err != nil {
		t.Errorf("Encountered error %s by StdoutNonBlocking()", err.Error())
	}
	serr, err := programstate.StderrNonBlocking()
	if err != nil {
		t.Errorf("Encountered error %s by StderrNonBlocking()", err.Error())
	} else if len(serr) != 0 {
		fmt.Println("Returned error by StderrNonBlocking()", string(serr))
	}

	want := "Hello"
	got := string(out)

	if got != want {
		t.Errorf("Got text = %q, want %q", got, want)
	}

	//send an interrupt signal to blocked binary
	err = programstate.Command.Process.Signal(os.Interrupt)
	if err != nil {
		t.Errorf("Encountered error %s while sending interrupt signal", err.Error())
	}

	out, err = programstate.StdoutNonBlocking()
	if err != nil {
		t.Errorf("Encountered error %s by StdoutNonBlocking()", err.Error())
	}

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
