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
func Setup(ctxCh chan context.Context) (*xgb.Conn, xproto.Window, *QuitStruct, error) {

	q := new(QuitStruct)
	n := 1000 + rand.Intn(9999-1000+1) // the display number

	var xephyrLayout Layout
	X, screenInfo, err := Newconn(xephyrLayout, n, q)
	if err != nil {
		return nil, 0, nil, err
	}

	debuggingport1 := 9222
	debuggingport2 := 9223

	chromewindow1, chromewindow2, inputwindow := DefaultWindowsLayout(screenInfo)

	go CreateChromeWindow(chromewindow1, "/tmp/aso_sxs_viewer/dir1", ":"+strconv.Itoa(n), ForceQuit, X, screenInfo, q, debuggingport1, ctxCh)
	go CreateChromeWindow(chromewindow2, "/tmp/aso_sxs_viewer/dir2", ":"+strconv.Itoa(n), ForceQuit, X, screenInfo, q, debuggingport2, ctxCh)

	return CreateInputWindow(inputwindow, X, screenInfo, q)
}

// NewConn opens a Xephyr window on a particular display and connects to it
func Newconn(layout Layout, display int, a *QuitStruct) (*xgb.Conn, *xproto.ScreenInfo, error) {
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
			"-ac",
			"-screen",
			fmt.Sprintf("%dx%d+%d+%d", layout.w, layout.h, layout.x, layout.y),
			"-br",
			"-reset",
			"-no-host-grab",
		},
	}
	programstate, err := command.ExecuteProgram(xephyr, cmdErrorHandler)
	if err != nil {
		return nil, nil, err
	}
	a.Quitters = append(a.Quitters, ChromeWindow{programstate})

	for {
		_, err = os.Stat("/tmp/.X11-unix/X" + strconv.Itoa(display))
		if !os.IsNotExist(err) {
			break
		}
	}

	// step2: start a connection with parent Xephyr on parent display
	X, err := xgb.NewConnDisplay(displayString)
	if err != nil {
		return nil, nil, err
	}

	setup := xproto.Setup(X)
	screenInfo := setup.DefaultScreen(X)

	return X, screenInfo, nil
}

// CreateChromeWindow opens chrome browser session in linux
func CreateChromeWindow(layout Layout, userdatadir string, display string, quitfunc func(*QuitStruct),
	X *xgb.Conn, screenInfo *xproto.ScreenInfo, a *QuitStruct, debuggingPort int, ctxCh chan context.Context) error {

	chromewindow := command.ExternalCommand{
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

		Env: []string{
			"DISPLAY=" + display},
	}

	programstate, err := command.ExecuteProgram(chromewindow, cmdErrorHandler)
	if err != nil {
		log.Println("Could not execute google-chrome command. Encountered error %s", err.Error())
		return err
	}

	ctx, err := establishChromeConnection(programstate, CHROMECONNTIMEOUT)
	if err != nil {
		log.Println(err)
		return err
	}

	ctxCh <- ctx

	a.Quitters = append(a.Quitters, ChromeWindow{programstate})

	// Close everything in case Chrome stops working
	for {
		if programstate.IsRunning() == false {
			fmt.Println("chrome closed- calling force quit")
			quitfunc(a)
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
}
