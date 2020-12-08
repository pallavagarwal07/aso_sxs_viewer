package createwindow

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"sync"

	"github.com/chromedp/chromedp"
	"github.com/pallavagarwal07/aso_sxs_viewer/command"
	"github.com/pallavagarwal07/aso_sxs_viewer/config"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/randr"
	"github.com/jezek/xgb/xproto"
)

const (
	chromeConnTimeout = 30
)

// Session contains all information that will be needed by main.
type Session struct {
	mu                   sync.Mutex
	X                    *xgb.Conn
	InputWin             InputWindow
	RootWin              RootWindow
	ChromeList           []ChromeWindow
	browserInputBarFocus Focus
}

type ChromeWindow struct {
	progState          *command.ProgramState
	Ctx                context.Context
	InputFieldSelector CSSSelector
}
type InputWindow struct {
	Wid  xproto.Window
	Conn *xgb.Conn
}
type RootWindow struct {
	progState *command.ProgramState
}
type CSSSelector struct {
	Selector string
	Position int
}

type Focus struct {
	isFocussed bool
	mu         sync.Mutex
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
func (p *RootWindow) Quit() {
	p.progState.Command.Process.Kill()
}

func (p *ChromeWindow) Quit() {
	p.progState.Command.Process.Kill()
}

func (p *RootWindow) ToClose() bool {
	return p.progState.IsRunning()
}

func (p *ChromeWindow) ToClose() bool {
	return p.progState.IsRunning()
}

func (p *InputWindow) Quit() {
	p.Conn.Close()
}

func (s *Session) SetBrowserInputBarFocus(isFocussed bool) {
	s.browserInputBarFocus.mu.Lock()
	defer s.browserInputBarFocus.mu.Unlock()
	s.browserInputBarFocus.isFocussed = isFocussed
}

func (s *Session) GetBrowserInputBarFocus() bool {
	s.browserInputBarFocus.mu.Lock()
	defer s.browserInputBarFocus.mu.Unlock()
	isfocussed := s.browserInputBarFocus.isFocussed
	return isfocussed
}

func (s *Session) InitializeChromeWindows(browserList []*config.BrowserConfig, cmdList []command.ExternalCommand, cmdErrorHandler func(err error) error) error {
	for i := 0; i < len(browserList); i++ {
		go s.initializeChromeWindow(browserList[i], cmdList[i], cmdErrorHandler)
	}
	return nil
}

func (s *Session) initializeChromeWindow(browserConfig *config.BrowserConfig, cmd command.ExternalCommand, cmdErrorHandler func(err error) error) error {
	chromeWindow, err := CreateChromeWindow(browserConfig, cmd, cmdErrorHandler)
	if err != nil {
		log.Println(err)
		return err
	}

	if err := SetupChrome(chromeWindow, browserConfig.GetUrl()); err != nil {
		log.Println(err)
		return err
	}
	s.appendChromeList(chromeWindow)
	return nil
}

// CreateChromeWindow opens a Chrome browser session.
func CreateChromeWindow(browserConfig *config.BrowserConfig, cmd command.ExternalCommand, cmdErrorHandler func(err error) error) (ChromeWindow, error) {

	programstate, err := command.ExecuteProgram(cmd, cmdErrorHandler)
	if err != nil {
		return ChromeWindow{}, err
	}

	ctx, err := establishChromeConnection(programstate, chromeConnTimeout)
	if err != nil {
		return ChromeWindow{}, err
	}

	selector := CSSSelector{browserConfig.GetCssSelector().GetSelector(), int(browserConfig.GetCssSelector().GetPosition())}
	return ChromeWindow{programstate, ctx, selector}, nil
}

func SetupChrome(chromeWindow ChromeWindow, URL string) error {
	if err := chromedp.Run(chromeWindow.Ctx, chromedp.Navigate(URL)); err != nil {
		return err
	}
	return nil
}
func DisableCrashedBubble(s string) error {
	fileinfo, err := os.Stat(s)
	if os.IsNotExist(err) {
		return nil
	}
	read, err := ioutil.ReadFile(s)
	if err != nil {
		return err
	}
	newContents := strings.Replace(string(read), `"exit_type":"Crashed"`, `"exit_type":"Normal"`, -1)
	return ioutil.WriteFile(s, []byte(newContents), fileinfo.Mode())
}

// Layout has the x , y coordinates of top left corner and width and height of window.
type Layout struct {
	x, y uint32
	w, h uint16
}

func WindowsLayout(screenInfo *xproto.ScreenInfo, inputOrientation string, n int) (chromeLayouts []Layout, inputwindow Layout) {
	var yShift uint32
	heightScreen := 0.85 * float64(screenInfo.HeightInPixels)
	widthScreen := screenInfo.WidthInPixels
	inputwindow.h, inputwindow.w = uint16(0.15*float64(screenInfo.HeightInPixels)), uint16(widthScreen)

	if inputOrientation == "TOP" {
		inputwindow.y = 0
		yShift = uint32(inputwindow.h)
	} else {
		inputwindow.y = uint32(heightScreen)
		yShift = 0
	}

	rows := int(n/4) + 1
	columns := int(math.Ceil(float64(n) / float64(rows)))

	var temp Layout
	temp.h = uint16(int(heightScreen) / rows)
	temp.w = uint16(int(widthScreen) / columns)

	for i, r := 0, 0; r < rows; r++ {
		temp.y = uint32(uint16(r)*temp.h) + yShift
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
	if s.RootWin.ToClose() {
		s.RootWin.Quit()
	}
}

func (s *Session) CreateXephyrWindow(layout Layout, display int, cmdErrorHandler func(err error) error) error {
	displayString := fmt.Sprintf(":%d", display)
	xephyr := command.ExternalCommand{
		Path: "Xephyr",
		Arg: []string{
			displayString,
			"-ac",
			"-screen", fmt.Sprintf("%dx%d+%d+%d", layout.w, layout.h, layout.x, layout.y),
			"-br",
			"-reset",
			"-no-host-grab",
		},
	}
	programstate, err := command.ExecuteProgram(xephyr, cmdErrorHandler)
	if err != nil {
		return err
	}
	s.RootWin = RootWindow{programstate}

	for {
		_, err = os.Stat("/tmp/.X11-unix/X" + strconv.Itoa(display))
		if !os.IsNotExist(err) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

func DefaultRootWindowSize() (height, width uint16) {
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
func Setup(viewerConfig *config.ViewerConfig) (*Session, error) {
	debuggingport := 9222
	var displayString string
	browserCount := viewerConfig.GetBrowserCount()

	var session Session

	// calls ForceQuit in case ChromeWindows are closed.
	cmdErrorHandler := func(err error) error {
		if err != nil {
			fmt.Printf("returned error %s, calling force quit", err.Error())
		}
		session.ForceQuit()
		return err
	}

	if runtime.GOOS != "darwin" {
		displayNumber := 1000 + rand.Intn(9999-1000+1)
		displayString = fmt.Sprintf(":%d", displayNumber)
		var rootLayout Layout
		rootLayout.x, rootLayout.y, rootLayout.w, rootLayout.h = viewerConfig.GetRootWindowLayout()
		if err := session.CreateXephyrWindow(rootLayout, displayNumber, cmdErrorHandler); err != nil {
			return nil, err
		}
	}

	screenInfo, err := session.Newconn(displayString)
	if err != nil {
		return nil, err
	}

	inputOrientation := viewerConfig.GetInputWindowOrientation()
	chromeLayouts, inputWindowLayout := WindowsLayout(screenInfo, inputOrientation, browserCount)

	var cmdList []command.ExternalCommand
	userDataDirPath := viewerConfig.GetUserDataDirPath()
	for i := 1; i <= browserCount; i++ {
		cmd := ChromeCommand(chromeLayouts[i-1], fmt.Sprintf("%s/dir%d", userDataDirPath, i), displayString, debuggingport+i)
		cmdList = append(cmdList, cmd)
		if err = DisableCrashedBubble(fmt.Sprintf("%s/dir%d/Default/Preferences", userDataDirPath, i)); err != nil {
			fmt.Println(err)
		}
	}

	browserList := viewerConfig.GetBrowserConfigList()
	if err := session.InitializeChromeWindows(browserList, cmdList, cmdErrorHandler); err != nil {
		return nil, err
	}

	if err := session.CreateInputWindow(inputWindowLayout, session.X, screenInfo); err != nil {
		return nil, err
	}
	return &session, nil
}
