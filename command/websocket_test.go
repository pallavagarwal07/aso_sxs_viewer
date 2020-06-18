package command

import (
	"os"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
)

func TestWsURL(t *testing.T) {

	runfilePath := "command/tests/sample_program_websocket_test/sample_program_websocket_test_/sample_program_websocket_test"
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

	want := "ws://127.0.0.1:9222/devtools/browser/ce8f8213-2323-4a88-9924-6b15247213e1"
	got, err := WsURL(programstate)
	if err != nil {
		t.Errorf("Encountered error %s in WsURL()", err.Error())
	}
	if got != want {
		t.Errorf("Got text = %q, want %q", got, want)
	}

	//test for err in case ws:// does not exist
	runfilePath = "command/tests/sample_program_websocket_test_inf/sample_program_websocket_test_inf_/sample_program_websocket_test_inf"
	testSampleCodePath, err = bazel.Runfile(runfilePath)
	if err != nil {
		t.Errorf("Encountered error %s by bazel.Runfile with arg %s", err.Error(), runfilePath)
	}
	test = ExternalCommand{
		Path: testSampleCodePath,
		Env:  os.Environ(),
	}

	programstate, err = ExecuteProgram(test, testErrorHandler)
	if err != nil {
		t.Errorf("Encountered error %s by ExecuteProgram()", err.Error())
	}

	if _, err := WsURL(programstate); err.Error() != "websocket url timeout reached" {
		t.Errorf("Unexpected behavior in case when ws:// does not exist")
	}

}
