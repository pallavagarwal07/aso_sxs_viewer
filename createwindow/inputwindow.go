package main

import (
	"fmt"
	"log"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

//CreateInputWindow creates window to capture keycodes
func CreateInputWindow(x uint32, y uint32, w uint16, h uint16, quitfunc func(*QuitStruct),
	X *xgb.Conn, screenInfo *xproto.ScreenInfo, a *QuitStruct) (*xgb.Conn, xproto.Window) {

	wid, _ := xproto.NewWindowId(X)
	xproto.CreateWindow(X, screenInfo.RootDepth, wid, screenInfo.Root,
		0, 0, w, h, 0,
		xproto.WindowClassInputOutput, screenInfo.RootVisual,
		xproto.CwBackPixel|xproto.CwEventMask,
		[]uint32{
			0xffffffff,
			xproto.EventMaskStructureNotify |
				xproto.EventMaskKeyPress |
				xproto.EventMaskKeyRelease})
	xproto.MapWindow(X, wid)
	xproto.ConfigureWindow(X, wid,
		xproto.ConfigWindowX|xproto.ConfigWindowY,
		[]uint32{
			y, x,
		})
	cookie := xproto.ChangeWindowAttributesChecked(
		X, wid, xproto.CwEventMask,
		[]uint32{
			xproto.EventMaskPointerMotion |
				xproto.EventMaskButtonPress |
				xproto.EventMaskButtonRelease |
				xproto.EventMaskEnterWindow |
				xproto.EventMaskLeaveWindow |
				xproto.EventMaskKeyPress |
				xproto.EventMaskStructureNotify |
				xproto.EventMaskKeyRelease})

	if err := cookie.Check(); err != nil {
		log.Fatalln(err)
	}

	(a.quitters) = append(a.quitters, InputWindow{wid, X, true})

	eventloop(X, wid, a, quitfunc)

	return X, wid
}

func eventloop(X *xgb.Conn, wid xproto.Window, a *QuitStruct, quitfunc func(*QuitStruct)) {
	for {

		ev, err := X.WaitForEvent()

		if err == nil && ev == nil {
			fmt.Println("connection interrupted")
			(a.quitters)[len(a.quitters)-1].SetToClose(false)
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
			// default:
			// fmt.Println(a)
		}

	}

}

func keyPresshandler(X *xgb.Conn, wid xproto.Window, a *QuitStruct, quitfunc func(*QuitStruct), b xproto.KeyPressEvent) {
	fmt.Println(b.Detail)
}

func keyReleasehandler(X *xgb.Conn, wid xproto.Window, a *QuitStruct, quitfunc func(*QuitStruct), b xproto.KeyReleaseEvent) {
	fmt.Println(b.Detail)
}

func enterNotifyhandler(X *xgb.Conn, wid xproto.Window, a *QuitStruct, quitfunc func(*QuitStruct), b xproto.EnterNotifyEvent) {
	cookie := xproto.GrabKeyboard(X, true, wid, xproto.TimeCurrentTime, xproto.GrabModeAsync, xproto.GrabModeAsync)
	if _, err := cookie.Reply(); err != nil {
		log.Fatalln(err)
	}
	// keybinding.Focus = false
}

func leaveNotifyhandler(X *xgb.Conn, wid xproto.Window, a *QuitStruct, quitfunc func(*QuitStruct), b xproto.LeaveNotifyEvent) {
	cookie := xproto.UngrabKeyboardChecked(X, xproto.TimeCurrentTime)
	if err := cookie.Check(); err != nil {
		log.Fatalln(err)
	}
}

func unmapNotifyhandler(X *xgb.Conn, wid xproto.Window, a *QuitStruct, quitfunc func(*QuitStruct), b xproto.UnmapNotifyEvent) {
	fmt.Println("unmap notify event")
	fmt.Println("connection interrupted")
	(a.quitters)[len(a.quitters)-1].SetToClose(false)
	quitfunc(a)
}
