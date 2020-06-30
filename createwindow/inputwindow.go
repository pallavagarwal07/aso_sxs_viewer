package main

import (
	"fmt"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

//CreateInputWindow creates window to capture keycodes
func CreateInputWindow(layout Layout, quitfunc func(*QuitStruct),
	X *xgb.Conn, screenInfo *xproto.ScreenInfo, a *QuitStruct) (*xgb.Conn, xproto.Window, error) {

	wid, _ := xproto.NewWindowId(X)
	cookie := xproto.CreateWindowChecked(X, screenInfo.RootDepth, wid, screenInfo.Root,
		0, 0, layout.w, layout.h, 0,
		xproto.WindowClassInputOutput, screenInfo.RootVisual,
		xproto.CwBackPixel|xproto.CwEventMask,
		[]uint32{
			0xffffffff,
			xproto.EventMaskEnterWindow |
				xproto.EventMaskLeaveWindow |
				xproto.EventMaskKeyPress |
				xproto.EventMaskKeyRelease |
				xproto.EventMaskStructureNotify})

	if err := cookie.Check(); err != nil {
		return nil, 0, err
	}

	xproto.MapWindow(X, wid)
	xproto.ConfigureWindow(X, wid,
		xproto.ConfigWindowX|xproto.ConfigWindowY,
		[]uint32{
			uint32(layout.y), uint32(layout.x),
		})

	a.quitters = append(a.quitters, InputWindow{wid, X, true})

	eventloop(X, wid, a, quitfunc)

	return X, wid, nil
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

func keyPresshandler(X *xgb.Conn, wid xproto.Window, a *QuitStruct, quitfunc func(*QuitStruct), b xproto.KeyPressEvent) {
	fmt.Println(b.Detail)
}

func keyReleasehandler(X *xgb.Conn, wid xproto.Window, a *QuitStruct, quitfunc func(*QuitStruct), b xproto.KeyReleaseEvent) {
	fmt.Println(b.Detail)
}

func enterNotifyhandler(X *xgb.Conn, wid xproto.Window, a *QuitStruct, quitfunc func(*QuitStruct), b xproto.EnterNotifyEvent) error {
	cookie := xproto.GrabKeyboard(X, true, wid, xproto.TimeCurrentTime, xproto.GrabModeAsync, xproto.GrabModeAsync)
	if _, err := cookie.Reply(); err != nil {
		return err
	}
	// keybinding.Focus = false
	return nil
}

func leaveNotifyhandler(X *xgb.Conn, wid xproto.Window, a *QuitStruct, quitfunc func(*QuitStruct), b xproto.LeaveNotifyEvent) error {
	cookie := xproto.UngrabKeyboardChecked(X, xproto.TimeCurrentTime)
	if err := cookie.Check(); err != nil {
		return err
	}
	return nil
}

func unmapNotifyhandler(X *xgb.Conn, wid xproto.Window, a *QuitStruct, quitfunc func(*QuitStruct), b xproto.UnmapNotifyEvent) {
	fmt.Println("unmap notify event")
	fmt.Println("connection interrupted")
	a.quitters[len(a.quitters)-1].SetToClose(false)
	quitfunc(a)
}
