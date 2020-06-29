// +build linux

package main

import (
	"fmt"
	"log"
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
	X, XChild, screenInfo, screenInfoChild := Newconn(0, 0, 1600, 1600, ":3", q)

	go CreateChromeWindow(0, 300, 600, 600, "/tmp/aso_sxs_viewer/dir1", ":3", ForceQuit, X, screenInfo, q)
	go CreateChromeWindow(650, 300, 600, 600, "/tmp/aso_sxs_viewer/dir2", ":3", ForceQuit, X, screenInfo, q)
	time.Sleep(5 * time.Second)
	CreateInputWindow(0, 0, 1280, 180, ForceQuit, XChild, screenInfoChild, q)
}

// NewConn opens a Xephyr window on a particular display and connects to it
func Newconn(x int, y int, w int, h int, display string, a *QuitStruct) (*xgb.Conn, *xgb.Conn, *xproto.ScreenInfo, *xproto.ScreenInfo) {
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

	for {
		_, err := os.Stat("/tmp/.X11-unix/X3")
		if !os.IsNotExist(err) {
			fmt.Println("File exists")
			break
		}
	}

	displayChild := ":4"

	xephyrChild := command.ExternalCommand{
		Path: "Xephyr",
		Arg: []string{
			displayChild,
			"-ac",
			"-screen",
			strconv.Itoa(w) + "x" + strconv.Itoa(200) + "+" + strconv.Itoa(x) + "+" + strconv.Itoa(y),
			"-br",
			"-reset",
		},
		Env: []string{
			"DISPLAY=" + display},
	}
	programstate, err = command.ExecuteProgram(xephyrChild, cmdErrorHandler)
	if err != nil {
		log.Fatal(err)
	}

	for {
		_, err := os.Stat("/tmp/.X11-unix/X4")
		if !os.IsNotExist(err) {
			fmt.Println("File exists")
			break
		}
	}

	// step2: start a connection with Xephyr on that particular display
	X, err := xgb.NewConnDisplay(display)
	if err != nil {
		log.Fatal(err)
	}

	// step2: start a connection with Xephyr on that particular display
	XChild, err := xgb.NewConnDisplay(displayChild)
	if err != nil {
		log.Fatal(err)
	}

	setup := xproto.Setup(X)
	screenInfo := setup.DefaultScreen(X)
	setupChild := xproto.Setup(XChild)
	screenInfoChild := setupChild.DefaultScreen(XChild)

	return X, XChild, screenInfo, screenInfoChild
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

	// Close everything in case Chrome stops working
	for {
		if programstate.IsRunning() == false {
			fmt.Println("chrome closed- calling force quit")
			myfunc(a)
			return
		}
	}
}
