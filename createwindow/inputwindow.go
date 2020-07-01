package createwindow

import (
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

//CreateInputWindow creates window to capture keycodes
func CreateInputWindow(layout Layout, X *xgb.Conn, screenInfo *xproto.ScreenInfo, a *QuitStruct) (*xgb.Conn, xproto.Window, *QuitStruct, error) {
	wid, _ := xproto.NewWindowId(X)
	cookie := xproto.CreateWindowChecked(X, screenInfo.RootDepth, wid, screenInfo.Root,
		0, 0, uint16(layout.w), uint16(layout.h), 0,
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
		return nil, 0, a, err
	}

	xproto.MapWindow(X, wid)
	xproto.ConfigureWindow(X, wid,
		xproto.ConfigWindowX|xproto.ConfigWindowY,
		[]uint32{
			uint32(layout.x), uint32(layout.y),
		})

	a.Quitters = append(a.Quitters, &InputWindow{wid, X})

	return X, wid, a, nil
}
