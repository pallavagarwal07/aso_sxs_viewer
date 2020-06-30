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

	X, wid, quitStruct, err := createwindow.Setup(ctxCh)
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
	eventLoop(X, wid, quitStruct, createwindow.ForceQuit)
}

func eventLoop(X *xgb.Conn, wid xproto.Window, a *createwindow.QuitStruct, quitfunc func(*createwindow.QuitStruct)) {
	for {

		ev, err := X.WaitForEvent()
		if err != nil {
			errorHandler(err)
			continue
		}

		if ev == nil {
			fmt.Println("connection interrupted")
			a.Quitters[len(a.Quitters)-1].SetToClose(false)
			quitfunc(a)
			return
		}

		switch e := ev.(type) {

		case xproto.KeyPressEvent:
			fmt.Println("keypress encountered")
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
			createwindow.UnmapNotifyHandler(a, quitfunc)
			return
		}

	}

}

func errorHandler(err error) {
	log.Println(err)
}

func Navigate(ctxList []context.Context, url string) error {
	for _, ctx := range ctxList {
		fmt.Println(ctx)
		if err := chromedp.Run(ctx,
			chromedp.Navigate(url),
		); err != nil {
			return err
		}
	}
	return nil
}