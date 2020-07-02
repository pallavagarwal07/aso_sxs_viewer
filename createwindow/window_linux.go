// +build linux

package createwindow

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"../command"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/randr"
	"github.com/jezek/xgb/xproto"
)

// ChromeCommand will take structured data after config file is implemented
func ChromeCommand(layout Layout, userdatadir, display string,
	debuggingPort int) command.ExternalCommand {
	cmd := command.ExternalCommand{
		Path: "google-chrome",
		Arg: []string{
			"--user-data-dir=" + userdatadir,
			fmt.Sprintf("--window-position=%d,%d", layout.x, layout.y),
			fmt.Sprintf("--window-size=%d,%d", layout.w, layout.h),
			"--disable-session-crashed-bubble",
			"--disble-infobars",
			"--disable-extensions",
			fmt.Sprintf("--remote-debugging-port=%d", debuggingPort),
		},
		Env: []string{"DISPLAY=" + display},
	}
	return cmd
}

// CreateXephyrWindow opens a Xephyr window on a particular display and connects to it
func (s *Session) CreateXephyrWindow(layout Layout, display int, cmdErrorHandler func(p *command.ProgramState, err error) error) error {
	if layout.h == 0 {
		layout.h = windowHeight
	}
	if layout.w == 0 {
		layout.w = windowWidth
	}

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
	s.appendChromeList(ChromeWindow{programstate})

	for {
		_, err = os.Stat("/tmp/.X11-unix/X" + strconv.Itoa(display))
		if !os.IsNotExist(err) {
			break
		}
	}
	return nil
}

func DefaultXephyrSize() (height, width uint16) {
	height, width = 900, 1600 // Sensible defaults in case the below fails.

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
