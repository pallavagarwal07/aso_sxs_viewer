// +build linux

package createwindow

// +build linux

package createwindow

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"../command"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// Setup opens all windows and establishes connection with the x server
func Setup(n int , ctxCh chan context.Context) (*xgb.Conn, xproto.Window, error) {
	debuggingport:= 9222
	var displayString string

	var session Session{}
	cmdErrorHandler := func(p *command.ProgramState, err error) error {
		if err != nil {
			fmt.Println("returned error %s, calling force quit", err.Error())
		}
		session.ForceQuit()
		return err
	}

	if runtime.GOOS != "darwin"{
	displayNumber := 1000 + rand.Intn(9999-1000+1)
	displayString := fmt.Sprintf(":%d",displayNumber)
	var xephyrLayout Layout
	xephyrLayout.h, xephyrLayout.w = DefaultXephyrSize()
	if err := CreateXephyrWindow(xephyrLayout,n,cmdErrorHandler);err != nil {
		return nil, 0, err
	}
	}

	X, screenInfo, err := newconn(fmt.Sprintf(":%d",n))
	if err != nil {
		return nil, 0, err
	}

	chromeLayouts, inputWindowLayout:= DefaultWindowsLayout(screenInfo,n)

	for i:= 1; i<=n ; i++{
		cmd := ChromeCommand(chromeLayouts[i-1],fmt.Sprintf("%s/.aso_sxs_viewer/profiles/dir%d",os.Getenv("HOME"),i),displayString,debuggingport+i)
		go session.CreateChromeWindow(cmd,ctxCh,cmdErrorHandler)
	}

	return CreateInputWindow(inputwindow, X, screenInfo)
}

// NewConn opens a Xephyr window on a particular display and connects to it
func CreateXephyrWindow(layout Layout, display int,cmdErrorHandler func( *command.ProgramState, err error) error)  error{
	// step1: start xephyr on a particular display number with position and size
	if layout.h == 0 {
		layout.h = WINDOWHEIGHT
	}
	if layout.w == 0 {
		layout.w = WINDOWWIDTH
	}

	displayString := fmt.Sprintf(":%d", display)
	xephyr := command.ExternalCommand{
		Path: "Xephyr",
		Arg: []string{
			displayString,
			"-ac","-screen",
			fmt.Sprintf("%dx%d+%d+%d", layout.w, layout.h, layout.x, layout.y),
			"-br",
			"-reset",
			"-no-host-grab",
		},
	}
	programstate, err := command.ExecuteProgram(xephyr, cmdErrorHandler)
	if err != nil {
		return err
	}
	s.appendWindowList(ChromeWindow{programstate})

	for {
		_, err = os.Stat("/tmp/.X11-unix/X" + strconv.Itoa(display))
		if !os.IsNotExist(err) {
			break
		}
	}
	return nil
}

// PopulateCommand will take structured data after config file is implemented
func ChromeCommand(layout Layout, userdatadir, display string,
	debuggingPort int) command.ExternalCommand {
    cmd := command.ExternalCommand{
		Path: "google-chrome",
		Arg: []string{
			"--user-data-dir=" + userdatadir,
			fmt.Sprintf("--window-position=%d,%d", layout.x, layout.y),
			fmt.Sprintf("--window-size=%d,%d", layout.w, layout.h),
			"--disable-session-crashed-bubble",
			"--disble-infobars",
			"--disable-extensions",
			fmt.Sprintf("--remote-debugging-port=%d", debuggingPort),
		},
        Env: []string{ "DISPLAY=" + display},
	}
	return cmd
}
