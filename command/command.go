package command

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"sync"
)

// ExternalCommand has the properties for invoking exec.Command
type ExternalCommand struct {
	Path string
	Arg  []string
	Env  []string
}

type ProgramState struct {
	Command     *exec.Cmd
	stdoutPipe  io.ReadCloser
	stderrPipe  io.ReadCloser
	stdout      []byte
	stderr      []byte
	stdoutmutex sync.Mutex
	stderrmutex sync.Mutex
}

// ExecuteProgram invokes an external program
func ExecuteProgram(command ExternalCommand, errorHandler func(*ProgramState, error) error) (*ProgramState, error) {
	var err error
	programState := &ProgramState{}
	cmd := exec.Command(command.Path, command.Arg...)
	cmd.Env = append(os.Environ(), command.Env...)
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

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	go func() {
		programState.StdoutListner()
		waitGroup.Done()
	}()

	programState.StderrListner()
	waitGroup.Wait()

	err := programState.Command.Wait()
	errorHandler(programState, err)

}

func (p *ProgramState) StdoutListner() {

	Listner(p.stdoutPipe, &p.stdout, &p.stdoutmutex)

}

func (p *ProgramState) StderrListner() {

	Listner(p.stderrPipe, &p.stderr, &p.stderrmutex)
}

func Listner(pipe io.ReadCloser, buffer *[]byte, mutex *sync.Mutex) {

	reader := bufio.NewReader(pipe)
	for {

		buf, err := reader.ReadBytes('\n')
		mutex.Lock()
		*buffer = append(*buffer, buf...)
		mutex.Unlock()
		if err == io.EOF {
			break
		}
	}

}

func (p *ProgramState) IsRunning() bool {
	if p.Command.ProcessState == nil {
		return true
	}
	return !p.Command.ProcessState.Exited()
}

func (p *ProgramState) Stdout() []byte {
	p.stdoutmutex.Lock()
	defer p.stdoutmutex.Unlock()
	stdout := make([]byte, len(p.stdout))
	copy(stdout, p.stdout)
	return stdout
}

func (p *ProgramState) Stderr() []byte {
	p.stderrmutex.Lock()
	defer p.stderrmutex.Unlock()
	stderr := make([]byte, len(p.stderr))
	copy(stderr, p.stderr)
	return stderr
}
