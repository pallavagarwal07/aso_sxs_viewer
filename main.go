package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/pallavagarwal07/aso_sxs_viewer/config"
	"github.com/pallavagarwal07/aso_sxs_viewer/createwindow"
	"github.com/pallavagarwal07/aso_sxs_viewer/event"

	"github.com/jezek/xgb/xproto"
)

func main() {
	rand.Seed(time.Now().Unix())

	viewerConfig, err := config.GetConfig()
	if err != nil {
		event.DisplayFatalError(err)
	}

	session, err := createwindow.Setup(viewerConfig)
	if err != nil {
		event.DisplayFatalError(err)
		return
	}

	if err := event.MapNotifyHandler(session); err != nil {
		event.DisplayFatalError(err)
		return
	}

	eventLoop(session)
}

func eventLoop(session *createwindow.Session) {
	for {
		ev, err := session.X.WaitForEvent()
		if err != nil {
			event.DisplayWarning(err)
			continue
		}

		if ev == nil {
			event.DisplayFatalError(fmt.Errorf("connection interrupted"))
			session.ForceQuit()
			return
		}

		switch e := ev.(type) {

		case xproto.KeyPressEvent:
			event.KeyPressHandler(session, &e)

		case xproto.MapNotifyEvent:
			if err := event.MapNotifyHandler(session); err != nil {
				event.DisplayFatalError(err)
				return
			}

		case xproto.EnterNotifyEvent:
			if err := event.EnterNotifyHandler(session); err != nil {
				event.DisplayWarning(err)
			}

		case xproto.LeaveNotifyEvent:
			if err := event.LeaveNotifyHandler(session); err != nil {
				event.DisplayWarning(err)
			}

		case xproto.UnmapNotifyEvent:
			event.UnmapNotifyHandler(session)
			return
		}
	}
}
