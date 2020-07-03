package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"./createwindow"
	"./keybinding"

	"github.com/chromedp/chromedp"
	"github.com/jezek/xgb/xproto"
)

const (
	chromeWindowNumber = 5
	URL                = "https://mail.google.com"
)

func main() {
	rand.Seed(time.Now().Unix())
	// var openChromeWin int

	session, err := createwindow.Setup(chromeWindowNumber)
	if err != nil {
		errorHandler(err)
		return
	}

	/*for openChromeWin < chromeWindowNumber {
			session.BrowserList = append(session.BrowserList, ctx)
	        openChromeWin++
		}*/

	/*if err := Navigate(session.BrowserList, URL); err != nil {
		errorHandler(err)
		return
	}*/

	if err := Navigate(session.ChromeList, URL); err != nil {
		errorHandler(err)
		return
	}

	if err := keybinding.UpdateMaps(session.X); err != nil {
		errorHandler(err)
		return
	}

	eventLoop(session)
}

func eventLoop(session *createwindow.Session) {
	for {
		ev, err := session.X.WaitForEvent()
		if err != nil {
			errorHandler(err)
			continue
		}

		if ev == nil {
			fmt.Println("connection interrupted")
			session.ForceQuit()
			return
		}

		switch ev.(type) {

		case xproto.KeyPressEvent:
			errorHandler(err)

		case xproto.MapNotifyEvent:
			errorHandler(err)

		case xproto.EnterNotifyEvent:
			errorHandler(err)

		case xproto.LeaveNotifyEvent:
			errorHandler(err)

		case xproto.UnmapNotifyEvent:
			createwindow.UnmapNotifyHandler(session.ForceQuit)
			return
		}
	}
}

func errorHandler(err error) {
	fmt.Println("error handler is to be implemented")
	log.Println(err)
}

func Navigate(chromeList []createwindow.ChromeWindow, url string) error {
	for _, chrome := range chromeList {
		ctx := chrome.Ctx
		go func(ctx context.Context) {
			if err := chromedp.Run(ctx,
				chromedp.Navigate(url),
			); err != nil {
				errorHandler(err)
			}
		}(ctx)
	}
	return nil
}

/*func Navigate(chromeList []createwindow.ChromeWindow, url string) error {
	for _, chrome := range chromeList {
		go func(chrome createwindow.ChromeWindow) {
			if err := chromedp.Run(chrome.Ctx,
				chromedp.Navigate(url),
			); err != nil {
				errorHandler(err)
			}
		}(chrome)
	}
	return nil
}*/
