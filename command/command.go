package command

import (
	"io"
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
func ExecuteProgram(command ExternalCommand, errorHandler func(*ProgramState, error) error) (*ProgramState, error) {
	var err error
	programState := &ProgramState{}
	cmd := exec.Command(command.Path, command.Arg...)
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

	go CloseProgram(programState, errorHandler)

	return programState, err
}

func CloseProgram(programState *ProgramState, errorHandler func(*ProgramState, error) error) {

	_, err := programState.StderrNonBlocking()
	if err != nil {
		errorHandler(programState, err)
		return
	}
	_, err = programState.StdoutNonBlocking()
	if err != nil {
		errorHandler(programState, err)
		return
	}

	err = programState.Command.Wait()
	errorHandler(programState, err)

}

func (p *ProgramState) StdoutNonBlocking() ([]byte, error) {

	return NonBlockingCall(p.stdoutPipe, p.stdout)

}

func (p *ProgramState) StderrNonBlocking() ([]byte, error) {

	return NonBlockingCall(p.stderrPipe, p.stderr)
}

func NonBlockingCall(pipe io.ReadCloser, buffer []byte) ([]byte, error) {

	var output []byte
	n, err := pipe.Read(output)
	if n > 0 && err == nil {
		buffer = append(buffer, output[:n]...)
	} else if n == 0 && err == io.EOF {
		err = nil
	}
	return buffer, err
}

func (p *ProgramState) isRunning() bool {
	return !p.Command.ProcessState.Exited()
}
