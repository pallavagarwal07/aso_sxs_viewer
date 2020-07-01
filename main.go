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
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

const CHROMEWINDOWNUM = 2
const URL = "https://mail.google.com"

func main() {
	rand.Seed(time.Now().Unix())
	ctxCh := make(chan context.Context)
	var openChromeWin int

	X, wid, err := createwindow.Setup(ctxCh)
	if err != nil {
		errorHandler(err)
		return
	}

	for openChromeWin < CHROMEWINDOWNUM {
		ctx := <-ctxCh
		keybinding.BrowserList = append(keybinding.BrowserList, ctx)
		openChromeWin++
	}

	if err := Navigate(keybinding.BrowserList, URL); err != nil {
		errorHandler(err)
		return
	}

	if err := keybinding.UpdateMaps(X); err != nil {
		errorHandler(err)
		return
	}
	eventLoop(X, wid, createwindow.ForceQuit)
}

func eventLoop(X *xgb.Conn, wid xproto.Window, quitfunc func()) {
	for {

		ev, err := X.WaitForEvent()
		if err != nil {
			errorHandler(err)
			continue
		}

		if ev == nil {
			fmt.Println("connection interrupted")
			quitfunc()
			return
		}

		switch e := ev.(type) {

		case xproto.KeyPressEvent:
			if err := keybinding.KeyPressHandler(X, keybinding.KeyPressEvent{&e}); err != nil {
				errorHandler(err)
			}
		case xproto.MapNotifyEvent:
			if err := keybinding.UpdateMaps(X); err != nil {
				errorHandler(err)
				return
			}
		case xproto.EnterNotifyEvent:
			keybinding.IsFocussed.SetFocus(false)
			if err := createwindow.EnterNotifyHandler(X, wid); err != nil {
				errorHandler(err)
			}
		case xproto.LeaveNotifyEvent:
			if err := createwindow.LeaveNotifyHandler(X); err != nil {
				errorHandler(err)
			}
		case xproto.UnmapNotifyEvent:
			createwindow.UnmapNotifyHandler(quitfunc)
			return
		}

	}

}

func errorHandler(err error) {
	log.Println(err)
}

func Navigate(ctxList []context.Context, url string) error {
	for _, ctx := range ctxList {
		if err := chromedp.Run(ctx,
			chromedp.Navigate(url),
		); err != nil {
			return err
		}
	}
	return nil
}
