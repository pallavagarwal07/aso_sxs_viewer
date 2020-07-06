package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/googleinterns/aso_sxs_viewer/config"
	"github.com/googleinterns/aso_sxs_viewer/createwindow"
	"github.com/googleinterns/aso_sxs_viewer/event"

	"github.com/jezek/xgb/xproto"
)

const (
	chromeWindowNumber = 2
)

func main() {
	rand.Seed(time.Now().Unix())

	viewerConfig, err := config.GetConfig()
	if err != nil {
		event.ErrorHandler(err)
	}

	session, err := createwindow.Setup(viewerConfig)
	if err != nil {
		event.ErrorHandler(err)
		return
	}

	if err := event.MapNotifyHandler(session); err != nil {
		event.ErrorHandler(err)
		return
	}

	eventLoop(session)
}

func eventLoop(session *createwindow.Session) {
	for {
		ev, err := session.X.WaitForEvent()
		if err != nil {
			event.ErrorHandler(err)
			continue
		}

		if ev == nil {
			fmt.Println("connection interrupted")
			session.ForceQuit()
			return
		}

		switch e := ev.(type) {

		case xproto.KeyPressEvent:
			event.KeyPressHandler(session, &e)

		case xproto.MapNotifyEvent:
			if err := event.MapNotifyHandler(session); err != nil {
				event.ErrorHandler(err)
				return
			}

		case xproto.EnterNotifyEvent:
			if err := event.EnterNotifyHandler(session); err != nil {
				event.ErrorHandler(err)
			}

		case xproto.LeaveNotifyEvent:
			if err := event.LeaveNotifyHandler(session); err != nil {
				event.ErrorHandler(err)
			}

		case xproto.UnmapNotifyEvent:
			event.UnmapNotifyHandler(session)
			return
		}
	}
}
