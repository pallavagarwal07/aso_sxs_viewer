package main

import (
	"fmt"
	"strconv"

	"../command"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

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
