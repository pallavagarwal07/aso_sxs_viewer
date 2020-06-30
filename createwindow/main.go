package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func main() {
	rand.Seed(time.Now().Unix())
	Setup()
}

func eventloop(X *xgb.Conn, wid xproto.Window, a *QuitStruct, quitfunc func(*QuitStruct)) {
	for {

		ev, err := X.WaitForEvent()

		if err == nil && ev == nil {
			fmt.Println("connection interrupted")
			a.quitters[len(a.quitters)-1].SetToClose(false)
			quitfunc(a)
			return
		}

		switch b := ev.(type) {

		case xproto.KeyPressEvent:
			keyPresshandler(X, wid, a, quitfunc, b)
		case xproto.KeyReleaseEvent:
			keyReleasehandler(X, wid, a, quitfunc, b)
		case xproto.EnterNotifyEvent:
			enterNotifyhandler(X, wid, a, quitfunc, b)
		case xproto.LeaveNotifyEvent:
			leaveNotifyhandler(X, wid, a, quitfunc, b)
		case xproto.UnmapNotifyEvent:
			unmapNotifyhandler(X, wid, a, quitfunc, b)
		}

	}

}
