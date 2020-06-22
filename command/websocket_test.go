package command

import (
	"os"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
)

func TestWsURL(t *testing.T) {
	//wsURL timeout in seconds
	timeout := 1

	runfilePath := "command/testdata/mock_chrome/mock_chrome_/mock_chrome"
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
	got, err := WsURL(programstate, timeout)
	if err != nil {
		t.Errorf("Encountered error %s in WsURL()", err.Error())
	}
	if got != want {
		t.Errorf("Got text = %q, want %q", got, want)
	}

	//test for err in case ws:// does not exist
	runfilePath = "command/testdata/infinite_loop/infinite_loop_/infinite_loop"
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

	if _, err := WsURL(programstate, timeout); err.Error() != "websocket url timeout reached" {
		t.Errorf("Unexpected behavior in case when ws:// does not exist")
	}
}
