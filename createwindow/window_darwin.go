// +build darwin

package main

import (
	"fmt"
	"time"

	"../command"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// Setup opens all the windows and establishes connection with the X server
func Setup() {
	q := new(QuitStruct)
	X, screenInfo := Newconn()

	go CreateChromeWindow(layout1, "/tmp/aso_sxs_viewer/dir1", ForceQuit, X, screenInfo, q)
	go CreateChromeWindow(layout2, "/tmp/aso_sxs_viewer/dir2", ForceQuit, X, screenInfo, q)
	CreateInputWindow(layout3, ForceQuit, X, screenInfo, q)
}

// Newconn establishes connection with XQuartz
func Newconn() (*xgb.Conn, *xproto.ScreenInfo, error) {
	X, err := xgb.NewConn()

	if err != nil {
		return nil, nil, err
	}

	setup := xproto.Setup(X)
	screenInfo := setup.DefaultScreen(X)
	return X, screenInfo, nil
}

// CreateChromeWindow opens a Chrome browser session
func CreateChromeWindow(layout Layout, userdatadir string, quitfunc func(*QuitStruct),
	X *xgb.Conn, screenInfo *xproto.ScreenInfo, a *QuitStruct) {

	cmd := command.ExternalCommand{
		Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		Arg: []string{"--user-data-dir=" + userdatadir,
			fmt.Sprintf("--window-position=%d,%d", layout.x, layout.y),
			fmt.Sprintf("--window-position=%d,%d", layout.w, layout.h),
			"--disable-session-crashed-bubble", "--disble-infobars", "--disable-extensions"},
	}

	programstate, err := command.ExecuteProgram(cmd, cmdErrorHandler)

	if err != nil {
		return err
	}

	a.quitters = append(a.quitters, ChromeWindow{programstate})

	for {
		time.Sleep(10 * time.Millisecond)
		if programstate.IsRunning() == false {
			fmt.Println("chrome closed- calling force quit")
			quitfunc(a)
			return
		}
	}
}
