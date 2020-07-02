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
	chromeWindowNumber = 5 // upto 4 is working, 5 is not - don't know why
	URL                = "https://mail.google.com"
)

func main() {
	rand.Seed(time.Now().Unix())
	ctxCh := make(chan context.Context)
	var openChromeWin int

	session, err := createwindow.Setup(chromeWindowNumber, ctxCh)
	if err != nil {
		errorHandler(err)
		return
	}

	for openChromeWin < chromeWindowNumber {
		ctx := <-ctxCh
		session.BrowserList = append(session.BrowserList, ctx)
		openChromeWin++
	}

	if err := Navigate(session.BrowserList, URL); err != nil {
		errorHandler(err)
		return
	}

	// session.X vs. session.inputWin.Conn??
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

func Navigate(ctxList []context.Context, url string) error {
	for _, ctx := range ctxList {
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
