package command

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
)

func TestExecuteProgram(t *testing.T) {

	runfilePath := "command/tests/tests_/tests"
	testSampleCodePath, err := bazel.Runfile(runfilePath)
	if err != nil {
		t.Errorf("Encountered error %s by bazel.Runfile with arg %s", err.Error(), runfilePath)
	}
	test := ExternalCommand{
		Path: testSampleCodePath,
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
