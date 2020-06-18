package main

import (
	"fmt"
	"strconv"

	"sync"

	"aso_sxs_viewer/command"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

var wg sync.WaitGroup

//Close has method quit that closes that window and kills that program
type Close interface {
	Quit()
}

//XquartzWindow is
type XquartzWindow struct {
	Wid    xproto.Window
	Conn   *xgb.Conn
	IsOpen bool
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
	p.Conn.Close()
}

//this has to be changed: need to pass two programstates and XQuartzwindow to this 
func cmdErrorHandler(p *command.ProgramState, err error) error {
	//programstate.Command.Process.Kill()
	//programstate2.Command.Process.Kill()
	return err
}

//ForceQuit closes everything in case of any error
func ForceQuit(p []*command.ProgramState, w XquartzWindow) {
	var mutex = &sync.Mutex{}

	mutex.Lock()
	fmt.Println("starting force quit")

	//if programstate.IsRunning() == true {

	p1 := ChromeWindow{p[0]}
	p1.Quit()
	fmt.Println("first chrome quit")
	//}

	//if programstate2.IsRunning() == true {

	p2 := ChromeWindow{p[1]}
	p2.Quit()
	fmt.Println("second chrome quit")
	//}

	if w.IsOpen == true {
		w.Quit()
		fmt.Println("window quit")
	}

	mutex.Unlock()

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
func Openbrowsersessions(X *xgb.Conn, screenInfo *xproto.ScreenInfo, programstateslice chan []*command.ProgramState) {

	heightScreen := screenInfo.HeightInPixels
	widthScreen := screenInfo.WidthInPixels

	h := int(heightScreen - 150)
	w := int(widthScreen / 2)

	cmd := command.ExternalCommand{
		Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		Arg: []string{"--window-position=0,0",
			"--window-size=" + strconv.Itoa(w) + "," + strconv.Itoa(h)},
	}

	var err error
	programstate, err := command.ExecuteProgram(cmd, cmdErrorHandler)

	if err != nil {
		fmt.Println(err)
	}

	cmd2 := command.ExternalCommand{
		Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		Arg: []string{"--user-data-dir=/Users/aditibhattacharya/chrome-dev-profile",
			"--window-position=" + strconv.Itoa(w) + ",0",
			"--window-size=" + strconv.Itoa(w) + "," + strconv.Itoa(h)},
	}

	programstate2, err := command.ExecuteProgram(cmd2, cmdErrorHandler)

	if err != nil {
		fmt.Println(err)
	}

	p := []*command.ProgramState{programstate, programstate2}
	programstateslice <- p

	wg.Done()

	command.CloseProgram(programstate2, cmdErrorHandler)
	command.CloseProgram(programstate, cmdErrorHandler)

}

//Createinputwindow creates window to caprure keycodes
func Createinputwindow(X *xgb.Conn, screenInfo *xproto.ScreenInfo) XquartzWindow {

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
	return XquartzWindow{wid, X, true}
}

//Capturekeys returns keycodes
func Capturekeys(X *xgb.Conn, p []*command.ProgramState, w XquartzWindow) {
	for {
		ev, err := X.WaitForEvent()

		if ev != nil && ev.Bytes()[0] == 2 {
			fmt.Println("yes, keypress or keyrelease, keycode:")
			fmt.Println(ev.Bytes()[1]) //prints the keycode.
		}
		if err == nil && ev == nil {
			fmt.Println("connection interrupted")
			w.IsOpen = false
			ForceQuit(p, w)
			return
		}
	}
}

func main() {
	X, screenInfo := Newconn()

	programstateslice := make(chan []*command.ProgramState)
	wg.Add(1)

	go Openbrowsersessions(X, screenInfo, programstateslice)

	p := <-programstateslice
	wg.Wait()
	close(programstateslice)

	X2, screenInfo := Newconn()
	w := Createinputwindow(X2, screenInfo)
	Capturekeys(X2, p, w)
}
