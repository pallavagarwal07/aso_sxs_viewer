package command

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestExecuteProgram(t *testing.T) {

	test := ExternalCommand{
		Path: "./test_command/test_sample_program",
		Env:  os.Environ(),
	}

	programstate, err := ExecuteProgram(test, testErrorHandler)
	if err != nil {
		t.Errorf("Encountered error %s by ExecuteProgram()", err.Error())
	}

	want := "Hello\n"
	got := testWaitOutput(programstate.Stdout, want)

	serr := programstate.Stderr()
	if len(serr) != 0 {
		fmt.Println("Returned error by StderrNonBlocking()", string(serr))
	}

	if got != want {
		t.Errorf("Got text = %q, want %q", got, want)
	}

	err = programstate.Command.Process.Signal(os.Interrupt)
	if err != nil {
		t.Errorf("Encountered error %s while sending interrupt signal", err.Error())
	}

	want = "Hello\nWorld\n"
	got = testWaitOutput(programstate.Stdout, want)

	serr = programstate.Stderr()
	if len(serr) != 0 {
		fmt.Println("Returned error by StderrNonBlocking()", string(serr))
	}

	if got != want {
		t.Errorf("Got text = %q, want %q", got, want)
	}
}

func testErrorHandler(p *ProgramState, err error) error {
	fmt.Println("Test_error_handler", err)
	return err
}

func testWaitOutput(getter func() []byte, want string) string {
	timeout := time.After(1 * time.Second)
	result := ""
	for result != want {
		select {
		case <-timeout:
			return result
		default:
		}
		result = string(getter())
	}

	return result
}
