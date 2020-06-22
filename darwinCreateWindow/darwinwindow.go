package main

import (
	"fmt"
	"strconv"

	"sync"

	"aso_sxs_viewer/command"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

//Close has method quit that closes that window and kills that program
type Quitters interface {
	Quit()
	ToClose() bool
}

//XquartzWindow is struct to hold information about the input window
type XquartzWindow struct {
	Wid    xproto.Window
	Conn   *xgb.Conn
	IsOpen bool
}

//ChromeWindow is struct to hold information about Chrome browser sessions
type ChromeWindow struct {
	*command.ProgramState
}

//Quit method to close the Chrome browser sessions
func (p ChromeWindow) Quit() {
	p.Command.Process.Kill()
}

//ToClose method checks whether ChromeWindow needs to be closed
func (p ChromeWindow) ToClose() bool {
	return !p.Command.ProcessState.Exited()
}

//Quit method to close the Xquartz (input) window
func (p XquartzWindow) Quit() {
	p.Conn.Close()
}

//ToClose method checks whether XquartzWindow needs to be closed
func (p XquartzWindow) ToClose() bool {
	return p.IsOpen
}

/*this has to be omitted*/
func cmdErrorHandler(p *command.ProgramState, err error) error {
	return err
}

//ForceQuit closes everything in case of any error
func ForceQuit(quitters *[]Quitters) {
	var mutex = &sync.Mutex{}

	mutex.Lock()
	fmt.Println("starting force quit")

	if (*quitters)[len(*quitters)-1].ToClose() == true {
		(*quitters)[len(*quitters)-1].Quit()
	}

	for _, q := range (*quitters)[:len(*quitters)-1] {
		if q.ToClose() == true {
			q.Quit()
		}
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
func CreateChromeWindow(x int, y int, w int, h int, s string, myfunc func(a *[]Quitters),
	X *xgb.Conn, screenInfo *xproto.ScreenInfo, quitters *[]Quitters) {

	cmd := command.ExternalCommand{
		Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		Arg: []string{"--user-data-dir=" + s,
			"--window-position=" + strconv.Itoa(x) + "," + strconv.Itoa(y),
			"--window-size=" + strconv.Itoa(w) + "," + strconv.Itoa(h)},
	}

	programstate, err := command.ExecuteProgram(cmd, cmdErrorHandler)

	if err != nil {
		fmt.Println(err)
	}

	(*quitters) = append(*quitters, ChromeWindow{programstate})

	for {
		if (*quitters)[0].ToClose() == false {
			myfunc(quitters)
		}
	}
}

//Createinputwindow creates window to caprure keycodes
func CreateInputWindow(x uint32, y uint32, w uint16, h uint16, myfunc func(a *[]Quitters),
	X *xgb.Conn, screenInfo *xproto.ScreenInfo, quitters *[]Quitters) {

	wid, _ := xproto.NewWindowId(X)
	xproto.CreateWindow(X, screenInfo.RootDepth, wid, screenInfo.Root,
		0, 0, w, h, 0,
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
			y, x,
		})

	(*quitters) = append(*quitters, XquartzWindow{wid, X, true})

	for {
		ev, err := X.WaitForEvent()

		if ev != nil && ev.Bytes()[0] == 2 {
			fmt.Println("yes, keypress or keyrelease, keycode:")
			fmt.Println(ev.Bytes()[1])
		}

		if err == nil && ev == nil {
			fmt.Println("connection interrupted")
			//how do i check if the connection has been interrupted if i can't update IsOpen
			//(*quitters)[len(*quitters)-1].ToClose() = false
			myfunc(quitters)
			return
		}
	}
}

func main() {
	X, screenInfo := Newconn()
	X2, screenInfo := Newconn()

	var myslice []Quitters
	var quitters *[]Quitters //pointer to a slice of quitters

	quitters = &myslice

	go CreateChromeWindow(0, 0, 600, 600, "/Users/aditibhattacharya/chrome-dev-profile", ForceQuit, X, screenInfo, quitters)

	CreateInputWindow(0, 0, 1280, 50, ForceQuit, X2, screenInfo, quitters)

}
