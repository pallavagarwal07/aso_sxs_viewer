package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"./createwindow"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func main() {
	rand.Seed(time.Now().Unix())
	ctxCh := make(chan context.Context)
	X, wid, quitStruct, err := createwindow.Setup(ctxCh)
	if err != nil {
		errorHandler(err)
	}

	eventLoop(X, wid, quitStruct, createwindow.ForceQuit)
}

func eventLoop(X *xgb.Conn, wid xproto.Window, a *createwindow.QuitStruct, quitfunc func(*createwindow.QuitStruct)) {
	for {

		ev, err := X.WaitForEvent()

		if err == nil && ev == nil {
			fmt.Println("connection interrupted")
			a.Quitters[len(a.Quitters)-1].SetToClose(false)
			quitfunc(a)
			return
		}

		switch b := ev.(type) {

		case xproto.KeyPressEvent:
			// keyPresshandler(X, wid, a, quitfunc, b)
		case xproto.KeyReleaseEvent:
			// keyReleasehandler(X, wid, a, quitfunc, b)
		case xproto.EnterNotifyEvent:
			// enterNotifyhandler(X, wid, a, quitfunc, b)
		case xproto.LeaveNotifyEvent:
			// leaveNotifyhandler(X, wid, a, quitfunc, b)
		case xproto.UnmapNotifyEvent:
			// unmapNotifyhandler(X, wid, a, quitfunc, b)
		}

	}

}

func errorHandler(err error) {
	log.Fatal(err)
}
