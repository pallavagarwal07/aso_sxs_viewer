package createwindow

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"

	"sync"

	"../command"

	"github.com/chromedp/chromedp"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

const (
	windowHeight      = 1400
	windowWidth       = 2000
	chromeConnTimeout = 30
)

//Session contains all information that will be needed by main
type Session struct {
	lock        sync.Mutex
	X           *xgb.Conn
	BrowserList []context.Context
	InputWin    InputWindow
	chromeList  []ChromeWindow
}

type ChromeWindow struct {
	*command.ProgramState
}

type InputWindow struct {
	Wid  xproto.Window
	Conn *xgb.Conn // see when this can be replaced with Session.X
}

func (s *Session) appendChromeList(chromeWin ChromeWindow) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.chromeList = append(s.chromeList, chromeWin)
}

func (s *Session) getChromeList() []ChromeWindow {
	s.lock.Lock()
	defer s.lock.Unlock()
	programs := make([]ChromeWindow, len(s.chromeList))
	copy(programs, s.chromeList)
	return programs
}

// Quit method to close the Chrome browser sessions
func (p ChromeWindow) Quit() {
	p.Command.Process.Kill()
}

// ToClose method checks whether ChromeWindow needs to be closed
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

// CreateChromeWindow opens a Chrome browser session
func (s *Session) CreateChromeWindow(cmd command.ExternalCommand, ctxCh chan context.Context,
	cmdErrorHandler func(p *command.ProgramState, err error) error) error {
	programstate, err := command.ExecuteProgram(cmd, cmdErrorHandler)
	if err != nil {
		log.Println(err)
		return err
	}

	ctx, err := establishChromeConnection(programstate, chromeConnTimeout)
	if err != nil {
		log.Println(err)
		return err
	}

	ctxCh <- ctx
	// s.appendWindowList(ChromeWindow{programstate})
	s.appendChromeList(ChromeWindow{programstate})
	return nil
}

// Layout has the x , y coordinates of top left corner and width and height of window
type Layout struct {
	x, y uint32
	w, h uint16
}

// WindowsLayout stores window size and position
func WindowsLayout(screenInfo *xproto.ScreenInfo, n int) (chromeLayouts []Layout, inputwindow Layout) {
	heightScreen := 0.8 * float64(screenInfo.HeightInPixels)
	widthScreen := screenInfo.WidthInPixels
	inputwindow.h, inputwindow.w = uint16(0.2*float64(screenInfo.HeightInPixels)), uint16(widthScreen)
	inputwindow.y = uint32(heightScreen)

	rows := int(n/4) + 1
	columns := uint16(math.Ceil(float64(n / rows)))

	var temp Layout
	temp.h = uint16(int(heightScreen) / rows)
	temp.w = uint16(widthScreen / columns)

	for i, r := 0, rows; r > 0; r-- {
		temp.y = uint32(uint16(r-1) * temp.h)
		fmt.Println(temp.y)
		for c := columns; c > 0 && i < n; c-- {
			temp.x = uint32((c - 1) * temp.w)
			fmt.Println(temp.x)
			chromeLayouts = append(chromeLayouts, temp)
			i++
		}
	}
	return chromeLayouts, inputwindow
}

// Setup opens all windows and establishes connection with the x server
func Setup(n int, ctxCh chan context.Context) (*Session, error) {
	debuggingport := 9222
	var displayString string
	var session Session

	// calls ForceQuit in case ChromeWindows are closed
	cmdErrorHandler := func(p *command.ProgramState, err error) error {
		if err != nil {
			fmt.Println("returned error %s, calling force quit", err.Error())
		}
		session.ForceQuit()
		return err
	}

	if runtime.GOOS != "darwin" {
		displayNumber := 1000 + rand.Intn(9999-1000+1)
		displayString = fmt.Sprintf(":%d", displayNumber)
		var xephyrLayout Layout
		xephyrLayout.h, xephyrLayout.w = DefaultXephyrSize()
		if err := session.CreateXephyrWindow(xephyrLayout, displayNumber, cmdErrorHandler); err != nil {
			return nil, err
		}
	}

	screenInfo, err := session.Newconn(displayString)
	if err != nil {
		return nil, err
	}

	chromeLayouts, inputWindowLayout := WindowsLayout(screenInfo, n)
	for i := 1; i <= n; i++ {
		cmd := ChromeCommand(chromeLayouts[i-1], fmt.Sprintf("%s/.aso_sxs_viewer/profiles/dir%d", os.Getenv("HOME"), i), displayString, debuggingport+i)
		go session.CreateChromeWindow(cmd, ctxCh, cmdErrorHandler)
	}
	if err := session.CreateInputWindow(inputWindowLayout, session.X, screenInfo); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *Session) Newconn(displayString string) (*xproto.ScreenInfo, error) {
	var err error
	s.X, err = xgb.NewConnDisplay(displayString)
	if err != nil {
		return nil, err
	}

	setup := xproto.Setup(s.X)
	screenInfo := setup.DefaultScreen(s.X)

	return screenInfo, nil
}

func establishChromeConnection(programState *command.ProgramState, timeout int) (context.Context, error) {
	wsURL, err := command.WsURL(programState, timeout)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to the chrome window. Encountered error %s", err.Error())
	}

	if wsURL == "" {
		return nil, errors.New("must specify -devtools-ws-url")
	}

	allocatorContext, _ := chromedp.NewRemoteAllocator(context.Background(), wsURL)

	ctx, _ := chromedp.NewContext(allocatorContext)

	return ctx, nil
}

// ForceQuit closes all open windows
func (s *Session) ForceQuit() {
	programs := s.getChromeList()

	fmt.Println("starting force quit")

	for _, q := range programs {
		if q.ToClose() == true {
			q.Quit() // will be quitting the other open Chrome Windows
		}
	}
	if s.InputWin.ToClose() == true {
		s.InputWin.Quit()
	}
}
