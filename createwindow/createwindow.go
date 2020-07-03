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
	"strconv"

	"sync"

	"github.com/googleinterns/aso_sxs_viewer/command"

	"github.com/chromedp/chromedp"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/randr"
	"github.com/jezek/xgb/xproto"
)

const (
	chromeConnTimeout = 30
	// will be removed once config is used.
	URL = "https://mail.google.com"
)

// Session contains all information that will be needed by main.
type Session struct {
	mu         sync.Mutex
	X          *xgb.Conn
	InputWin   InputWindow
	ChromeList []ChromeWindow
}

type ChromeWindow struct {
	progState *command.ProgramState
	Ctx       context.Context
}

type InputWindow struct {
	Wid  xproto.Window
	Conn *xgb.Conn
}

func (s *Session) appendChromeList(chromeWin ChromeWindow) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ChromeList = append(s.ChromeList, chromeWin)
}

func (s *Session) getChromeList() []ChromeWindow {
	s.mu.Lock()
	defer s.mu.Unlock()
	programs := make([]ChromeWindow, len(s.ChromeList))
	copy(programs, s.ChromeList)
	return programs
}

func (p ChromeWindow) Quit() {
	p.progState.Command.Process.Kill()
}

func (p ChromeWindow) ToClose() bool {
	return p.progState.IsRunning()
}

func (p *InputWindow) Quit() {
	p.Conn.Close()
}

func (s *Session) InitializeChromeWindow(cmd command.ExternalCommand,
	cmdErrorHandler func(p *command.ProgramState, err error) error, URL string) error {
	chromeWindow, err := CreateChromeWindow(cmd, cmdErrorHandler)
	if err != nil {
		log.Println(err)
		return err
	}

	if err := SetupChrome(chromeWindow, URL); err != nil {
		log.Println(err)
		return err
	}

	s.appendChromeList(chromeWindow)
	return nil
}

// CreateChromeWindow opens a Chrome browser session.
func CreateChromeWindow(cmd command.ExternalCommand,
	cmdErrorHandler func(p *command.ProgramState, err error) error) (ChromeWindow, error) {

	programstate, err := command.ExecuteProgram(cmd, cmdErrorHandler)
	if err != nil {
		return ChromeWindow{}, err
	}

	ctx, err := establishChromeConnection(programstate, chromeConnTimeout)
	if err != nil {
		return ChromeWindow{}, err
	}

	return ChromeWindow{programstate, ctx}, nil
}

func SetupChrome(chromeWindow ChromeWindow, URL string) error {
	if err := chromedp.Run(chromeWindow.Ctx, chromedp.Navigate(URL)); err != nil {
		return err
	}
	// TODO: use cookies to login if user allows.
	return nil
}

// Layout has the x , y coordinates of top left corner and width and height of window.
type Layout struct {
	x, y uint32
	w, h uint16
}

func WindowsLayout(screenInfo *xproto.ScreenInfo, n int) (chromeLayouts []Layout, inputwindow Layout) {
	heightScreen := 0.8 * float64(screenInfo.HeightInPixels)
	widthScreen := screenInfo.WidthInPixels
	inputwindow.h, inputwindow.w = uint16(0.2*float64(screenInfo.HeightInPixels)), uint16(widthScreen)
	inputwindow.y = uint32(heightScreen)

	rows := int(n/4) + 1
	columns := int(math.Ceil(float64(n) / float64(rows)))

	var temp Layout
	temp.h = uint16(int(heightScreen) / rows)
	temp.w = uint16(int(widthScreen) / columns)

	for i, r := 0, 0; r < rows; r++ {
		temp.y = uint32(uint16(r) * temp.h)
		for c := 0; c < columns && i < n; c++ {
			temp.x = uint32(uint16(c) * temp.w)
			chromeLayouts = append(chromeLayouts, temp)
			i++
		}
	}
	return chromeLayouts, inputwindow
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

// ForceQuit closes all open windows.
func (s *Session) ForceQuit() {
	programs := s.getChromeList()

	fmt.Println("starting force quit")

	for _, q := range programs {
		if q.ToClose() == true {
			q.Quit()
		}
	}
	// Input Window is gracefully closed, closing closed window is okay.
	s.InputWin.Quit()

}

func (s *Session) CreateXephyrWindow(layout Layout, display int, cmdErrorHandler func(p *command.ProgramState, err error) error) error {
	displayString := fmt.Sprintf(":%d", display)
	xephyr := command.ExternalCommand{
		Path: "Xephyr",
		Arg: []string{
			displayString,
			"-ac", "-screen",
			fmt.Sprintf("%dx%d+%d+%d", layout.w, layout.h, layout.x, layout.y),
			"-br",
			"-reset",
			"-no-host-grab",
		},
	}
	programstate, err := command.ExecuteProgram(xephyr, cmdErrorHandler)
	if err != nil {
		return err
	}
	// the Xephyr window is not a chromewindow but maintained in the same list.
	s.appendChromeList(ChromeWindow{programstate, nil})

	for {
		_, err = os.Stat("/tmp/.X11-unix/X" + strconv.Itoa(display))
		if !os.IsNotExist(err) {
			break
		}
	}
	return nil
}

func DefaultXephyrSize() (height, width uint16) {
	height, width = 900, 1600

	X, err := xgb.NewConn()
	if err != nil {
		log.Println(err)
		return
	}
	if err := randr.Init(X); err != nil {
		log.Println(err)
		return
	}

	screens, err := randr.GetScreenResourcesCurrent(X, xproto.Setup(X).DefaultScreen(X).Root).Reply()
	if err != nil {
		log.Println(err)
		return
	}

	crtc, err := randr.GetCrtcInfo(X, screens.Crtcs[0], xproto.TimeCurrentTime).Reply()
	if err != nil {
		log.Println(err)
		return
	}
	return uint16(0.8 * float64(crtc.Height)), uint16(0.8 * float64(crtc.Width))
}

// Setup opens all windows and establishes connection with the x server.
func Setup(n int) (*Session, error) {
	debuggingport := 9222
	var displayString string
	var session Session

	// calls ForceQuit in case ChromeWindows are closed.
	cmdErrorHandler := func(p *command.ProgramState, err error) error {
		if err != nil {
			fmt.Printf("returned error %s, calling force quit", err.Error())
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
		cmd := ChromeCommand(chromeLayouts[i-1], fmt.Sprintf("%s/.aso_sxs_viewer/profiles/dir%d", os.Getenv("HOME"), i),
			displayString, debuggingport+i)
		// Will take a BrowserConfig as paramenter later.
		go session.InitializeChromeWindow(cmd, cmdErrorHandler, URL)
	}
	if err := session.CreateInputWindow(inputWindowLayout, session.X, screenInfo); err != nil {
		return nil, err
	}
	return &session, nil
}
