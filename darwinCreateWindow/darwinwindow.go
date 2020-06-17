package main

import (
	"fmt"
	"strconv"

	"aso_sxs_viewer/command"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

var programstate, programstate2 *command.ProgramState

//Close has method quit that closes that window and kills that program
type Close interface {
	Quit()
}

//XquartzWindow is
type XquartzWindow struct {
	Wid  xproto.Window
	Conn *xgb.Conn
}

//ChromeWindow is
type ChromeWindow struct {
	*command.ProgramState
}

//Quit method to close the Chrome browser sessions
func (p ChromeWindow) Quit() {
	p.Command.Process.Kill()
}

//Quit method to close the Xquartz window
func (p XquartzWindow) Quit() {
	xproto.DestroyWindow(p.Conn, p.Wid)
}

//this has to be changed
func cmdErrorHandler(p *command.ProgramState, err error) error {
	programstate.Command.Process.Kill()
	programstate2.Command.Process.Kill()

	//including this line will quit all 3 windows (2 chrome and one Xquartz) the moment any of the browser sessions are quitted
	//log.Fatal("connection interrupted")

	return err
}

//ForceQuit closes everything in case of any error
func ForceQuit() {

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

//Openbrowsersessions opens two Chrome browser sessions side-by-side
func Openbrowsersessions(X *xgb.Conn, screenInfo *xproto.ScreenInfo) {
	heightScreen := screenInfo.HeightInPixels
	widthScreen := screenInfo.WidthInPixels

	h := int(heightScreen - 150)
	w := int(widthScreen / 2)

	fmt.Println(heightScreen, widthScreen)

	cmd := command.ExternalCommand{
		Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		Arg: []string{"--window-position=0,0",
			"--window-size=" + strconv.Itoa(w) + "," + strconv.Itoa(h)},
	}

	var err error
	programstate, err = command.ExecuteProgram(cmd, cmdErrorHandler)

	if err != nil {
		fmt.Println(err)
	}

	cmd2 := command.ExternalCommand{
		Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		Arg: []string{"--user-data-dir=/Users/aditibhattacharya/chrome-dev-profile",
			"--window-position=" + strconv.Itoa(w) + ",0",
			"--window-size=" + strconv.Itoa(w) + "," + strconv.Itoa(h)},
	}

	programstate2, err = command.ExecuteProgram(cmd2, cmdErrorHandler)

	if err != nil {
		fmt.Println(err)
	}

	command.CloseProgram(programstate2, cmdErrorHandler)
	command.CloseProgram(programstate, cmdErrorHandler)
}

//Createinputwindow creates window to caprure keycodes
func Createinputwindow(X *xgb.Conn, screenInfo *xproto.ScreenInfo) {

	heightScreen := screenInfo.HeightInPixels
	widthScreen := screenInfo.WidthInPixels

	h := uint32(heightScreen - 150)

	wid, _ := xproto.NewWindowId(X)
	xproto.CreateWindow(X, screenInfo.RootDepth, wid, screenInfo.Root,
		0, 0, widthScreen, 50, 0,
		xproto.WindowClassInputOutput, screenInfo.RootVisual,
		xproto.CwBackPixel|xproto.CwEventMask,
		[]uint32{
			0xffffffff,
			xproto.EventMaskStructureNotify |
				xproto.EventMaskKeyPress |
				xproto.EventMaskKeyRelease})
	xproto.MapWindow(X, wid)
	xproto.ConfigureWindow(X, wid,
		xproto.ConfigWindowX|xproto.ConfigWindowY,
		[]uint32{
			0, h,
		})
}

//Capturekeys returns keycodes
func Capturekeys(X *xgb.Conn) {

	ev, err := X.WaitForEvent()

	if ev.Bytes()[0] == 2 {
		fmt.Println("yes, keypress or keyrelease, keycode:")
	}
	fmt.Println(ev.Bytes()[1]) //prints the keycode.

	//this isn't doing anything atm
	/*if err == nil && ev==nil {
	fmt.Println("connection interrupted")
	ForceQuit()*/
}

func main() {
	X, screenInfo := Newconn()
	go Openbrowsersessions(X, screenInfo)
	X2, screenInfo := Newconn()
	Createinputwindow(X2, screenInfo)
	for {
		Capturekeys(X2)
	}
}
