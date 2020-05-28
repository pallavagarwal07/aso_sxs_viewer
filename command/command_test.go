package command

import (
	"fmt"
	"os"
	"testing"
)

func TestExecuteProgram(t *testing.T) {
	var outstream []byte
	var errstream []byte

	xephyr := ExternalCommand{
		path: "Xephyr",
		arg: []string{
			":3",
			"-ac",
			"-screen",
			"800x600",
			"-br",
			"-reset",
		},
		env: os.Environ(),
	}

	xterm := ExternalCommand{
		path: "xterm",
		env: append(
			os.Environ(),
			"DISPLAY=:3",
		),
	}

	c := make(chan Status)
	cx := make(chan Status)
	go ExecuteProgram(xephyr, outstream, errstream, c)

	if i := <-c; i.isRunning {
		go ExecuteProgram(xterm, outstream, errstream, cx)
		fmt.Println(string(errstream))
		fmt.Println(string(outstream))
	}

	if i := <-cx; i.err != nil {
		t.Errorf("Test failed %s", i.err.Error())
	}

}
