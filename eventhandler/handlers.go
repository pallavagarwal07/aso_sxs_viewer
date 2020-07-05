package eventhandler

import (
	"fmt"

	"github.com/jezek/xgb/xproto"
)

func EnterNotifyHandler(s *createwindow.Session) error {
	cookie := xproto.GrabKeyboard(s.X, true, s.InputWin, xproto.TimeCurrentTime, xproto.GrabModeAsync, xproto.GrabModeAsync)
	if _, err := cookie.Reply(); err != nil {
		return err
	}
	return nil
}

func LeaveNotifyHandler(s *createwindow.Session) error {
	cookie := xproto.UngrabKeyboardChecked(s.X, xproto.TimeCurrentTime)
	if err := cookie.Check(); err != nil {
		return err
	}
	return nil
}
func UnmapNotifyHandler(s *createwindow.Session) error {
	fmt.Println("Input window was closed,Connection interrupted")
	s.ForceQuit()
}
