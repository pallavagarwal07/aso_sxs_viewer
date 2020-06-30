// +build linux

package createwindow

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"../command"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// Setup opens all windows and establishes connection with the x server
func Setup() {

	q := new(QuitStruct)
	n := 1000 + rand.Intn(9999-1000+1) // the display number

	X, screenInfo := Newconn(0, 0, 1600, 1600, n, q)

	chromewindow1, chromewindow2, inputwindow := DefaultWindowsLayout(screenInfo)

	go CreateChromeWindow(chromewindow1, "/tmp/aso_sxs_viewer/dir1", ":"+strconv.Itoa(n), ForceQuit, X, screenInfo, q)
	go CreateChromeWindow(chromewindow2, "/tmp/aso_sxs_viewer/dir2", ":"+strconv.Itoa(n), ForceQuit, X, screenInfo, q)

	CreateInputWindow(inputwindow, ForceQuit, X, screenInfo, q)
}

// NewConn opens a Xephyr window on a particular display and connects to it
func Newconn(layout Layout, display int, a *QuitStruct) (*xgb.Conn, *xproto.ScreenInfo, error) {
	// step1: start xephyr on a particular display number with position and size

	fmt.Sprintf("%dx%d+%d+%d", w, h, x, y)

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
	a.quitters = append(a.quitters, ChromeWindow{programstate})

	for {
		_, err = os.Stat("/tmp/.X11-unix/X" + strconv.Itoa(display))
		if !os.IsNotExist(err) {
			fmt.Println("File exists")
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

	return nil, X, screenInfo
}

// CreateChromeWindow opens chrome browser session in linux
func CreateChromeWindow(layout Layout, userdatadir string, display string, quitfunc func(*QuitStruct),
	X *xgb.Conn, screenInfo *xproto.ScreenInfo, a *QuitStruct) error {

	chromewindow := command.ExternalCommand{
		Path: "google-chrome",
		Arg: []string{
			"--user-data-dir=" + userdatadir,
			fmt.Sprintf("--window-position=%d,%d", layout.x, layout.y),
			fmt.Sprintf("--window-position=%d,%d", layout.w, layout.h),
			"--disable-session-crashed-bubble", "--disble-infobars", "--disable-extensions"},

		Env: []string{
			"DISPLAY=" + display},
	}

	programstate, err := command.ExecuteProgram(chromewindow, cmdErrorHandler)
	if err != nil {
		return err
	}

	wsURL, err := command.WsURL(programstate)
	if err != nil {
		return err
	}

	a.quitters = append(a.quitters, ChromeWindow{programstate})

	// Close everything in case Chrome stops working
	for {
		if programstate.IsRunning() == false {
			fmt.Println("chrome closed- calling force quit")
			quitfunc(a)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
