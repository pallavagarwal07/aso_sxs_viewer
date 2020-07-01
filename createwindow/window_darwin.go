// +build darwin

package createwindow

import (
	"context"
	"fmt"
	"log"

	"../command"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// Setup opens all the windows and establishes connection with the X server
func Setup(ctxCh chan context.Context) (*xgb.Conn, xproto.Window, error) {
	X, screenInfo, err := Newconn()
	if err != nil {
		return nil, 0, err
	}

	chromewindow1, chromewindow2, inputwindow := DefaultWindowsLayout(screenInfo)

	debuggingport1 := 9222
	debuggingport2 := 9223

	go CreateChromeWindow(chromewindow1, "/tmp/aso_sxs_viewer/dir1", ForceQuit, X, screenInfo, debuggingport1, ctxCh)
	go CreateChromeWindow(chromewindow2, "/tmp/aso_sxs_viewer/dir2", ForceQuit, X, screenInfo, debuggingport2, ctxCh)
	return CreateInputWindow(inputwindow, X, screenInfo)
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
	X *xgb.Conn, screenInfo *xproto.ScreenInfo, debuggingport int, ctxCh chan context.Context) error {

	cmd := command.ExternalCommand{
		Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		Arg: []string{"--user-data-dir=" + userdatadir,
			fmt.Sprintf("--window-position=%d,%d", layout.x, layout.y),
			fmt.Sprintf("--window-size=%d,%d", layout.w, layout.h),
			"--disable-session-crashed-bubble", "--disble-infobars", "--disable-extensions",
			fmt.Sprintf("--remote-debugging-port=%d", debuggingport),
		},
	}

	programstate, err := command.ExecuteProgram(cmd, cmdErrorHandler)
	if err != nil {
		log.Println(err)
		return err
	}

	ctx, err := establishChromeConnection(programstate, CHROMECONNTIMEOUT)
	if err != nil {
		log.Println(err)
		return err
	}

	ctxCh <- ctx
	appendProgramList(ChromeWindow{programstate})
	//a.Quitters = append(a.Quitters, ChromeWindow{programstate})

	return nil
}
