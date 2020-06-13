package main

import (
	"fmt"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)


func main() {
	X, err := xgb.NewConn()
	if err != nil {
		fmt.Println(err)
		return
	}
	setup := xproto.Setup(X)
	screenInfo := setup.DefaultScreen(X)

	heightScreen := screenInfo.HeightInPixels
	widthScreen := screenInfo.WidthInPixels

	h := uint32(heightScreen - 150)

	wid, _ := xproto.NewWindowId(X)
	xproto.CreateWindow(X, screenInfo.RootDepth, wid, screenInfo.Root,
		0, 0, widthScreen, 50, 0,
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
			0, h,
		})
	for {
		ev, _ := X.WaitForEvent()
		if ev != nil {
			fmt.Println(ev.String())
			fmt.Println(ev.Bytes())
		}
		if ev.Bytes()[0] == 2 {
			fmt.Println("yes, keypress or keyrelease, keycode:")
		}
		fmt.Println(ev.Bytes()[1]) //prints the keycode.

	}

}
