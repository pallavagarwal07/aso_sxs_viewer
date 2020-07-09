package createwindow

import (
	"fmt"
	"math/rand"
	"runtime"

	"github.com/googleinterns/aso_sxs_viewer/command"
	"github.com/googleinterns/aso_sxs_viewer/config"
)

// Setup opens all windows and establishes connection with the x server and chrome.
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
