package command

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"time"
)

// WsURL returns the ws URL for the running browser window.
// Chrome will sometimes fail to print the websocket, or run for a long
// time, without properly exiting. To avoid blocking forever in those
// cases, give up after ten seconds.
func WsURL(programState *ProgramState, timeout int) (string, error) {
	wsURLReadTimeout := time.Duration(timeout) * time.Second
	var err error

	var wsURL string
	wsURLChan := make(chan struct{}, 1)
	go func() {
		wsURL, err = PollwsURL(programState)
		wsURLChan <- struct{}{}
	}()
	select {
	case <-wsURLChan:
	case <-time.After(wsURLReadTimeout):
		err = errors.New("websocket url timeout reached")
	}

	if err != nil {
		return wsURL, err
	}
	return wsURL, err
}

//PollwsURL retrives the wsURL from the Stderr of the process
func PollwsURL(programState *ProgramState) (string, error) {
	prefix := []byte("DevTools listening on")
	var wsURL string
	count := 0

	for {
		out := programState.Stderr()
		bytesReader := bytes.NewReader(out[count:])
		bufReader := bufio.NewReader(bytesReader)

		line, err := bufReader.ReadBytes('\n')

		if err == io.EOF {
			err = nil
		}
		if err != nil {
			return "", err
		}

		if bytes.HasPrefix(line, prefix) {
			line = line[len(prefix):]
			// use TrimSpace, to also remove \r on Windows
			line = bytes.TrimSpace(line)
			wsURL = string(line)
			break
		}
		count += len(line)
	}
	return wsURL, nil
}
