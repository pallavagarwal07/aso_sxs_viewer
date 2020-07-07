package event

import (
	"fmt"
	"log"

	"github.com/googleinterns/aso_sxs_viewer/chrometool"
	"github.com/googleinterns/aso_sxs_viewer/createwindow"
	"github.com/googleinterns/aso_sxs_viewer/keybinding"
	"github.com/jezek/xgb/xproto"
)

func KeyPressHandler(session *createwindow.Session, event *xproto.KeyPressEvent) error {
	str, mods := keybinding.InterpretKeyPressEvent(session.X, keybinding.KeyPressEvent{event})
	isFocussed := session.GetBrowserInputBarFocus()

	for _, browser := range session.ChromeList {
		go func(browser createwindow.ChromeWindow) {
			if err := chrometool.DispatchKeyEventToBrowser(browser.Ctx, chrometool.CSSSelector(browser.InputFieldSelector), str, mods, isFocussed); err != nil {
				ErrorHandler(err, false)
			}
		}(browser)
	}

	if str != "Return" {
		session.SetBrowserInputBarFocus(true)
	} else {
		session.SetBrowserInputBarFocus(false)
	}
	return nil
}

func MapNotifyHandler(session *createwindow.Session) error {
	if err := keybinding.UpdateMaps(session.X); err != nil {
		return err
	}
	return nil
}

func EnterNotifyHandler(session *createwindow.Session) error {
	session.SetBrowserInputBarFocus(false)
	cookie := xproto.GrabKeyboard(session.X, true, session.InputWin.Wid, xproto.TimeCurrentTime, xproto.GrabModeAsync, xproto.GrabModeAsync)
	if _, err := cookie.Reply(); err != nil {
		return err
	}
	xproto.ConfigureWindow(session.X, session.InputWin.Wid, xproto.ConfigWindowBorderWidth, []uint32{6})
	return nil
}

func LeaveNotifyHandler(session *createwindow.Session) error {
	cookie := xproto.UngrabKeyboardChecked(session.X, xproto.TimeCurrentTime)
	if err := cookie.Check(); err != nil {
		return err
	}
	xproto.ConfigureWindow(session.X, session.InputWin.Wid, xproto.ConfigWindowBorderWidth, []uint32{0})
	return nil
}

func UnmapNotifyHandler(session *createwindow.Session) {
	fmt.Println("Input window was closed,Connection interrupted")
	session.ForceQuit()
}

func ErrorHandler(err error, isFatal bool) {
	createwindow.DisplayError(err, PrintError, isFatal)
}

func PrintError(err error) error {
	if err != nil {
		log.Println(err)
	}
	return err
}
