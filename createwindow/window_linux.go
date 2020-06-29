// +build linux

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"../command"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

//Setup opens all windows and establishes connection with the x server
func Setup() {
	q := new(QuitStruct)
	X, screenInfo := Newconn(0, 0, 600, 600, ":3", q)

	go CreateChromeWindow(0, 0, 600, 600, "/tmp/aso_sxs_viewer/dir1", ":3", ForceQuit, X, screenInfo, q)
	CreateInputWindow(0, 0, 1280, 50, ForceQuit, X, screenInfo, q)
}

//NewConn opens a Xephyr window on a particular display and connects to it
func Newconn(x int, y int, w int, h int, display string, a *QuitStruct) (*xgb.Conn, *xproto.ScreenInfo) {
	// step1: start xephyr on a particular display number with position and size
	xephyr := command.ExternalCommand{
		Path: "Xephyr",
		Arg: []string{
			display,
			"-ac",
			"-screen",
			strconv.Itoa(w) + "x" + strconv.Itoa(h) + "+" + strconv.Itoa(x) + "+" + strconv.Itoa(y),
			"-br",
			"-reset",
		},
	}
	programstate, err := command.ExecuteProgram(xephyr, cmdErrorHandler)
	if err != nil {
		log.Fatal(err)
	}
	(a.quitters) = append(a.quitters, ChromeWindow{programstate})

	// step2: start a connection with Xephyr on that particular display
	X, err := xgb.NewConnDisplay(display)
	if err != nil {
		log.Fatal(err)
	}

	for {
		_, err := os.Stat("/tmp/.X11-unix/X3")
		if !os.IsNotExist(err) {
			fmt.Println("File exists")
			break
		}
	}

	setup := xproto.Setup(X)
	screenInfo := setup.DefaultScreen(X)
	return X, screenInfo
}

// CreateChromeWindow opens chrome browser session in linux
func CreateChromeWindow(x int, y int, w int, h int, userdatadir string, display string, myfunc func(*QuitStruct),
	X *xgb.Conn, screenInfo *xproto.ScreenInfo, a *QuitStruct) {

	chromewindow := command.ExternalCommand{
		Path: "google-chrome",
		Arg: []string{
			"--user-data-dir=" + userdatadir,
			"--window-position=" + strconv.Itoa(x) + "," + strconv.Itoa(y),
			"--window-size=" + strconv.Itoa(w) + "," + strconv.Itoa(h),
			"--disable-session-crashed-bubble", "--disble-infobars", "--disable-extensions"},

		Env: []string{
			"DISPLAY=" + display},
	}

	programstate, err := command.ExecuteProgram(chromewindow, cmdErrorHandler)
	if err != nil {
		log.Fatal(err)
	}

	(a.quitters) = append(a.quitters, ChromeWindow{programstate})

	for {
		ev, err := X.WaitForEvent()

		// Close everything in case Window is closed
		if ev != nil && ev.Bytes()[0] == xproto.UnmapNotify {
			fmt.Println("unmap notify event")
			fmt.Println("connection interrupted")
			(a.quitters)[len(a.quitters)-1].SetToClose(false)
			myfunc(a)
			return
		}

		// Close everything in case Chrome stops working
		if programstate.IsRunning() == false {
			fmt.Println("chrome closed- calling force quit")
			myfunc(a)
			return
		}
	}
}
