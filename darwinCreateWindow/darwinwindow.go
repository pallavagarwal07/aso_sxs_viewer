package main

import (
	"fmt"
	"strconv"

	"sync"

	"../command"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

//Quitters has method quit that closes that window and kills that program
type Quitters interface {
	Quit()
	ToClose() bool
	SetToClose(bool)
}

//XquartzWindow is struct to hold information about the input window
type XquartzWindow struct {
	Wid    xproto.Window
	Conn   *xgb.Conn
	IsOpen bool
}

//QuitStruct has the slice of quittes and the lock
type QuitStruct struct {
	quitters []Quitters
	lock     sync.Mutex
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
	return p.IsRunning()
}

//SetToClose method sets the value of IsRunning
func (p ChromeWindow) SetToClose(b bool) {
}

//Quit method to close the Xquartz (input) window
func (p XquartzWindow) Quit() {
	p.Conn.Close()
}

//ToClose method checks whether XquartzWindow needs to be closed
func (p XquartzWindow) ToClose() bool {
	return p.IsOpen
}

//SetToClose method sets the value of IsOpen
func (p XquartzWindow) SetToClose(b bool) {
	p.IsOpen = b
}

/*this has to be omitted*/
func cmdErrorHandler(p *command.ProgramState, err error) error {
	return err
}

//ForceQuit closes everything
func ForceQuit(a *QuitStruct) {

	a.lock.Lock()
	defer a.lock.Unlock()

	fmt.Println("starting force quit")

	if (a.quitters)[len(a.quitters)-1].ToClose() == true {
		(a.quitters)[len(a.quitters)-1].Quit()
		fmt.Println("quit Xwindow")
	}

	for _, q := range (a.quitters)[:len(a.quitters)-1] {
		if q.ToClose() == true {
			q.Quit() // will be quitting the other open Chrome Windows
			fmt.Println("quit Chrome window")
		}
	}

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
func CreateChromeWindow(x int, y int, w int, h int, s string, myfunc func(*QuitStruct),
	X *xgb.Conn, screenInfo *xproto.ScreenInfo, a *QuitStruct) {

	cmd := command.ExternalCommand{
		Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		Arg: []string{"--user-data-dir=" + s,
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

//CreateInputWindow creates window to caprure keycodes
func CreateInputWindow(x uint32, y uint32, w uint16, h uint16, myfunc func(*QuitStruct),
	X *xgb.Conn, screenInfo *xproto.ScreenInfo, a *QuitStruct) {

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

	(a.quitters) = append(a.quitters, XquartzWindow{wid, X, true})

	for {
		ev, err := X.WaitForEvent()

		if ev != nil && ev.Bytes()[0] == 2 {
			fmt.Println("yes, keypress or keyrelease, keycode:")
			fmt.Println(ev)
		}

		if err == nil && ev == nil {
			fmt.Println("connection interrupted")
			(a.quitters)[len(a.quitters)-1].SetToClose(false)
			myfunc(a)
			return
		}
	}
}

func main() {
	X, screenInfo := Newconn()
	X2, screenInfo := Newconn()

	q := new(QuitStruct)

	go CreateChromeWindow(0, 0, 600, 600, "/Users/aditibhattacharya/chrome-dev-profile", ForceQuit, X, screenInfo, q)
	CreateInputWindow(0, 0, 1280, 50, ForceQuit, X2, screenInfo, q)

}
