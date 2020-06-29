// +build darwin

package main

import (
	"fmt"
	"strconv"

	"../command"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

//Setup opens all the windows and establishes connection with the X server
func Setup() {
	q := new(QuitStruct)
	X, screenInfo := Newconn()

	go CreateChromeWindow(0, 0, 600, 600, "/tmp/aso_sxs_viewer/dir1", ForceQuit, X, screenInfo, q)
	CreateInputWindow(0, 0, 1280, 50, ForceQuit, X, screenInfo, q)
}

//Newconn establishes connection with XQuartz
func Newconn() (*xgb.Conn, *xproto.ScreenInfo) {
	X, err := xgb.NewConn()
	if err != nil {
		fmt.Println(err)
	}
	setup := xproto.Setup(X)
	screenInfo := setup.DefaultScreen(X)
	return X, screenInfo
}

//CreateChromeWindow opens a Chrome browser session
func CreateChromeWindow(x int, y int, w int, h int, userdatadir string, myfunc func(*QuitStruct),
	X *xgb.Conn, screenInfo *xproto.ScreenInfo, a *QuitStruct) {

	cmd := command.ExternalCommand{
		Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		Arg: []string{"--user-data-dir=" + userdatadir,
			"--window-position=" + strconv.Itoa(x) + "," + strconv.Itoa(y),
			"--window-size=" + strconv.Itoa(w) + "," + strconv.Itoa(h),
			"--disable-session-crashed-bubble", "--disble-infobars", "--disable-extensions"},
	}

	programstate, err := command.ExecuteProgram(cmd, cmdErrorHandler)

	if err != nil {
		fmt.Println(err)
	}

	(a.quitters) = append(a.quitters, ChromeWindow{programstate})

	for {
		if programstate.IsRunning() == false {
			fmt.Println("chrome closed- calling force quit")
			myfunc(a)
			return
		}
	}
}
