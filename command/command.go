package command

import (
	"io"
	"io/ioutil"
	"os/exec"
)

// ExternalCommand has the properties for invoking exec.Command
type ExternalCommand struct {
	Path string
	Arg  []string
	Env  []string
}

type ProgramState struct {
	Command    *exec.Cmd
	stdoutPipe io.ReadCloser
	stderrPipe io.ReadCloser
	stdout     []byte
	stderr     []byte
}

// ExecuteProgram invokes an external program
func ExecuteProgram(command ExternalCommand, errorHandler func(*exec.Cmd, error) error) (*ProgramState, error) {
	var err error
	programState := &ProgramState{}
	cmd := exec.Command(
		command.Path,
		command.Arg...)
	cmd.Env = command.Env
	programState.Command = cmd

	programState.stdoutPipe, err = cmd.StdoutPipe()
	if err != nil {
		return programState, err
	}

	programState.stderrPipe, err = cmd.StderrPipe()
	if err != nil {
		return programState, err
	}
	if err := cmd.Start(); err != nil {
		return programState, err
	}
	programState.StdoutNonBlocking()
	go CloseProgram(cmd, programState, errorHandler)

	return programState, err
}

func CloseProgram(cmd *exec.Cmd, programState *ProgramState, errorHandler func(*exec.Cmd, error) error) {

	programState.StderrNonBlocking()
	programState.StdoutNonBlocking()
	if err := cmd.Wait(); err != nil {
		errorHandler(cmd, err)
	}
}

func (p *ProgramState) StdoutNonBlocking() ([]byte, error) {

	stdout, err := ioutil.ReadAll(p.stdoutPipe)
	if err != nil {
		return p.stdout, err
	}

	p.stdout = append(p.stdout, stdout...)
	return p.stdout, err
}

func (p *ProgramState) StderrNonBlocking() ([]byte, error) {

	stderr, err := ioutil.ReadAll(p.stderrPipe)
	if err != nil {
		return p.stdout, err
	}

	p.stderr = append(p.stderr, stderr...)
	return p.stderr, err
}

func (p *ProgramState) isRunning() bool {
	return !p.Command.ProcessState.Exited()
}
