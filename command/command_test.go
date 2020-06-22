package command

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"time"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
)

func TestExecuteProgram(t *testing.T) {

	runfilePath := "command/testdata/ack_signal/ack_signal_/ack_signal"
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

	if err := programstate.Stderr(); len(err) != 0 {
		t.Errorf("Returned error %s by StderrNonBlocking()", string(err))
	}

	want := "Hello\n"
	if err := testWaitOutput(programstate.Stdout, want); err != nil {
		t.Errorf(err.Error())
	}

	if err := programstate.Command.Process.Signal(os.Interrupt); err != nil {
		t.Errorf("Encountered error %s while sending interrupt signal", err.Error())
	}

	if err := programstate.Stderr(); len(err) != 0 {
		t.Errorf("Returned error %s by StderrNonBlocking()", string(err))
	}

	want = "Hello\nWorld\n"
	if err := testWaitOutput(programstate.Stdout, want); err != nil {
		t.Errorf(err.Error())
	}
}

func testErrorHandler(p *ProgramState, err error) error {
	if err != nil {
		log.Fatalln("Invoked testErrorHandler with error %s", err.Error())
		return fmt.Errorf("Invoked testErrorHandler with error %s", err.Error())
	}
	return nil
}

func testWaitOutput(getter func() []byte, want string) error {
	timeout := time.After(1 * time.Second)
	result := ""
	for result != want {
		select {
		case <-timeout:
			return errors.New("Got text = " + result + " want " + want)
		default:
		}
		result = string(getter())
	}

	return nil
}
