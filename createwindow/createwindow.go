package createwindow

import (
	"context"
	"errors"
	"fmt"

	"sync"

	"../command"

	"github.com/chromedp/chromedp"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

const WINDOWHEIGHT = 1400
const WINDOWWIDTH = 2000
const CHROMECONNTIMEOUT = 30

// Layout has the x , y coordinates of top left corner and width and height of window
type Layout struct {
	x, y, w, h int
}

//Quitters has method quit that closes that window and kills that program
type Quitters interface {
	Quit()
	ToClose() bool
}

//InputWindow is struct to hold information about the input window
type InputWindow struct {
	Wid  xproto.Window
	Conn *xgb.Conn
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

// Quit method to close the input window
func (p *InputWindow) Quit() {
	p.Conn.Close()
}

// ToClose method checks whether InputWindow needs to be closed
func (p *InputWindow) ToClose() bool {
	return true
}

// tracks errors for execute program
func cmdErrorHandler(p *command.ProgramState, err error) error {
	fmt.Println(err)
	return err
}

//ForceQuit closes everything
func ForceQuit(a *QuitStruct) {

	a.lock.Lock()
	defer a.lock.Unlock()

	fmt.Println("starting force quit")

	for _, q := range a.Quitters {
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

	if wsURL == "" {
		return nil, errors.New("must specify -devtools-ws-url")
	}

	allocatorContext, _ := chromedp.NewRemoteAllocator(context.Background(), wsURL)

	// create context
	ctx, _ := chromedp.NewContext(allocatorContext)

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
