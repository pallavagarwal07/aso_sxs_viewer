package createwindow

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"sync"

	"github.com/googleinterns/aso_sxs_viewer/command"
	"github.com/googleinterns/aso_sxs_viewer/config"
	"github.com/jezek/xgb"
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

// Layout has the x , y coordinates of top left corner and width and height of window.
type Layout struct {
	x, y uint32
	w, h uint16
}

// WindowsLayout sets the position and sizes of windows to form a neat grid.
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

// DisableCrashedBubble sets Chrome exit type to Normal; no Restore Pages popup.
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
