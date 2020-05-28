package command

import (
	"bufio"
	"os/exec"
	"time"
)

// ExternalCommand has the properties for invoking exec.Command
type ExternalCommand struct {
	path string
	arg  []string
	env  []string
}

type Status struct {
	isRunning bool
	err       error
}

// ExecuteProgram invokes an external program
func ExecuteProgram(command ExternalCommand, stdout []byte, stderr []byte, status chan Status) error {
	var err error
	cmd := exec.Command(
		command.path,
		command.arg...)
	cmd.Env = command.env

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		status <- Status{
			isRunning: false,
			err:       err,
		}
		return err
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		status <- Status{
			isRunning: false,
			err:       err,
		}
		return err
	}

	stderrReader := bufio.NewReader(stderrPipe)
	stdoutReader := bufio.NewReader(stdoutPipe)

	if err := cmd.Start(); err != nil {
		return err
	}

	// TODO: check if client is ready
	time.Sleep(100 * time.Millisecond)
	status <- Status{
		isRunning: true,
		err:       err,
	}

	stderr, err = stderrReader.ReadBytes('\n')
	if err != nil {
		status <- Status{
			isRunning: false,
			err:       err,
		}
		return err
	}
	stdout, err = stdoutReader.ReadBytes('\n')
	if err != nil {
		status <- Status{
			isRunning: false,
			err:       err,
		}
		return err
	}

	if err := cmd.Wait(); err != nil {
		status <- Status{
			isRunning: false,
			err:       err,
		}
		return err
	}

	status <- Status{
		isRunning: false,
		err:       err,
	}
	close(status)
	return err
}
