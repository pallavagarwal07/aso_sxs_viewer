package command

import (
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
	Command *exec.Cmd
	stdout  []byte
	stderr  []byte
}

// ExecuteProgram invokes an external program
func ExecuteProgram(command ExternalCommand, errorHandler func(*exec.Cmd, error)) (*ProgramState, error) {
	var err error
	programState := &ProgramState{}
	cmd := exec.Command(
		command.Path,
		command.Arg...)
	cmd.Env = command.Env
	programState.Command = cmd

	if err := cmd.Start(); err != nil {
		errorHandler(cmd, err)
		return programState, err
	}

	if err := cmd.Wait(); err != nil {
		return programState, err
	}

	return programState, err
}

func (p *ProgramState) StdoutNonBlocking() ([]byte, error) {
	stdoutPipe, err := p.Command.StdoutPipe()
	if err != nil {
		return p.stdout, err
	}

	stdout, err := ioutil.ReadAll(stdoutPipe)
	if err != nil {
		return p.stdout, err
	}

	p.stdout = append(p.stdout, stdout...)
	return p.stdout, err
}

func (p *ProgramState) StderrNonBlocking() ([]byte, error) {
	stderrPipe, err := p.Command.StderrPipe()
	if err != nil {
		return p.stderr, err
	}

	stderr, err := ioutil.ReadAll(stderrPipe)
	if err != nil {
		return p.stdout, err
	}

	p.stderr = append(p.stderr, stderr...)
	return p.stderr, err
}

func (p *ProgramState) isRunning() bool {
	return !p.Command.ProcessState.Exited()
}
