// +build linux

package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"

	"../command"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// Setup opens all windows and establishes connection with the x server
func Setup() {

	q := new(QuitStruct)
	n := 1000 + rand.Intn(9999-1000+1) // the display number

	X, screenInfo := Newconn(0, 0, 1600, 1600, n, q)

	go CreateChromeWindow(0, 300, 600, 600, "/tmp/aso_sxs_viewer/dir1", ":"+strconv.Itoa(n), ForceQuit, X, screenInfo, q)
	go CreateChromeWindow(650, 300, 600, 600, "/tmp/aso_sxs_viewer/dir2", ":"+strconv.Itoa(n), ForceQuit, X, screenInfo, q)

	CreateInputWindow(0, 0, 1280, 180, ForceQuit, X, screenInfo, q)
}

// NewConn opens a Xephyr window on a particular display and connects to it
func Newconn(x int, y int, w int, h int, display int, a *QuitStruct) (*xgb.Conn, *xproto.ScreenInfo) {
	// step1: start xephyr on a particular display number with position and size

	displayString := ":" + strconv.Itoa(display)
	xephyr := command.ExternalCommand{
		Path: "Xephyr",
		Arg: []string{
			displayString,
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

	for {
		_, err := os.Stat("/tmp/.X11-unix/X" + strconv.Itoa(display))
		if !os.IsNotExist(err) {
			fmt.Println("File exists")
			break
		}
	}

	// step2: start a connection with parent Xephyr on parent display
	X, err := xgb.NewConnDisplay(displayString)
	if err != nil {
		log.Fatal(err)
	}

	setup := xproto.Setup(X)
	screenInfo := setup.DefaultScreen(X)

	return X, screenInfo
}

// CreateChromeWindow opens chrome browser session in linux
func CreateChromeWindow(x int, y int, w int, h int, userdatadir string, display string, quitfunc func(*QuitStruct),
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

	// Close everything in case Chrome stops working
	for {
		if programstate.IsRunning() == false {
			fmt.Println("chrome closed- calling force quit")
			quitfunc(a)
			return
		}
	}
}
