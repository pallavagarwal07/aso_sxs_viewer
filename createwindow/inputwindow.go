package main

import (
	"fmt"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

//CreateInputWindow creates window to capture keycodes
func CreateInputWindow(x uint32, y uint32, w uint16, h uint16, myfunc func(*QuitStruct),
	X *xgb.Conn, screenInfo *xproto.ScreenInfo, a *QuitStruct) {

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

	(a.quitters) = append(a.quitters, InputWindow{wid, X, true})

	for {
		ev, err := X.WaitForEvent()

		if ev != nil && ev.Bytes()[0] == 2 {
			fmt.Println("yes, keypress or keyrelease, keycode:")
			fmt.Println(ev)
		}

		if ev != nil && ev.Bytes()[0] == xproto.UnmapNotify {
			fmt.Println("unmap notify event")
			fmt.Println("connection interrupted")
			(a.quitters)[len(a.quitters)-1].SetToClose(false)
			myfunc(a)
			return
		}

		if err == nil && ev == nil {
			fmt.Println("connection interrupted")
			(a.quitters)[len(a.quitters)-1].SetToClose(false)
			myfunc(a)
			return
		}
	}
}
