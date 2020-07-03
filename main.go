package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/googleinterns/aso_sxs_viewer/createwindow"
	"github.com/googleinterns/aso_sxs_viewer/keybinding"

	"github.com/jezek/xgb/xproto"
)

const (
	chromeWindowNumber = 5
)

func main() {
	rand.Seed(time.Now().Unix())

	session, err := createwindow.Setup(chromeWindowNumber)
	if err != nil {
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
