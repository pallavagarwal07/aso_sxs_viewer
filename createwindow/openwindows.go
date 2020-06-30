package createwindow

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"sync"

	"../command"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/knq/chromedp"
)

const WINDOWHEIGHT = 700
const WINDOWWIDTH = 1200
const CHROMECONNTIMEOUT = 5

// Layout has the x , y coordinates of top left corner and width and height of window
type Layout struct {
	x, y, w, h int
}

//Quitters has method quit that closes that window and kills that program
type Quitters interface {
	Quit()
	ToClose() bool
	SetToClose(bool)
}

//InputWindow is struct to hold information about the input window
type InputWindow struct {
	Wid    xproto.Window
	Conn   *xgb.Conn
	IsOpen bool
}

//QuitStruct has the slice of quittes and the lock
type QuitStruct struct {
	Quitters []Quitters
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

//Quit method to close the input window
func (p InputWindow) Quit() {
	p.Conn.Close()
}

//ToClose method checks whether InputWindow needs to be closed
func (p InputWindow) ToClose() bool {
	return p.IsOpen
}

//SetToClose method sets the value of IsOpen
func (p InputWindow) SetToClose(b bool) {
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

	if (a.Quitters)[len(a.Quitters)-1].ToClose() == true {
		(a.Quitters)[len(a.Quitters)-1].Quit()
	}

	for _, q := range (a.Quitters)[:len(a.Quitters)-1] {
		if q.ToClose() == true {
			q.Quit() // will be quitting the other open Chrome Windows
		}
	}

}

func establishChromeConnection(programState *command.ProgramState, timeout int) (context.Context, error) {
	wsURL, err := command.WsURL(programState, timeout)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not connect to the chrome window. Encountered error %s", err.Error()))
	}

	var flagDevToolWsUrl = flag.String("devtools-ws-url", wsURL, "DevTools WebSsocket URL")
	flag.Parse()
	if *flagDevToolWsUrl == "" {
		return nil, errors.New("must specify -devtools-ws-url")
	}
	allocatorContext, cancel := chromedp.NewRemoteAllocator(context.Background(), *flagDevToolWsUrl)
	defer cancel()

	// create context
	ctx, cancel := chromedp.NewContext(allocatorContext)
	defer cancel()

	return ctx, nil
}

// DefaultWindowsLayout stores window size and position
func DefaultWindowsLayout(screenInfo *xproto.ScreenInfo) (chromewindow1, chromewindow2, inputwindow Layout) {

	heightScreen := screenInfo.HeightInPixels
	widthScreen := screenInfo.WidthInPixels

	chromewindow1.h = int(heightScreen - 150)
	chromewindow1.w = int(widthScreen / 2)
	chromewindow2.h = int(heightScreen - 150)
	chromewindow2.w = int(widthScreen / 2)
	inputwindow.h = 100
	inputwindow.w = int(widthScreen)

	chromewindow2.x = int(widthScreen / 2)
	inputwindow.y = int(heightScreen - 150)

	return chromewindow1, chromewindow2, inputwindow

}
