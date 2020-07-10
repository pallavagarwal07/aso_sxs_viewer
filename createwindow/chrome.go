package createwindow

import (
	"context"
	"errors"
	"fmt"

	"github.com/chromedp/chromedp"
	"github.com/googleinterns/aso_sxs_viewer/command"
	"github.com/googleinterns/aso_sxs_viewer/config"
)

// SetupChrome opens url in chrome broswer session
func SetupChrome(chromeWindow ChromeWindow, URL string) error {
	if err := chromedp.Run(chromeWindow.Ctx, chromedp.Navigate(URL)); err != nil {
		return err
	}
	return nil
}

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
