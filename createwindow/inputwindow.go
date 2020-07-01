package createwindow

import (
	"fmt"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// CreateInputWindow creates window to capture keycodes
func (s *Session) CreateInputWindow(layout Layout, X *xgb.Conn, screenInfo *xproto.ScreenInfo) (i InputWindow, err error) {
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
		return nil, err
	}

	xproto.MapWindow(X, wid)
	xproto.ConfigureWindow(X, wid,
		xproto.ConfigWindowX|xproto.ConfigWindowY,
		[]uint32{
			uint32(layout.x), uint32(layout.y),
		})

	s.appendWindowList(InputWindow{wid, X})

	return InputWindow{wid, X}, nil
}

func EnterNotifyHandler(X *xgb.Conn, wid xproto.Window) error {
	cookie := xproto.GrabKeyboard(X, true, wid, xproto.TimeCurrentTime, xproto.GrabModeAsync, xproto.GrabModeAsync)
	if _, err := cookie.Reply(); err != nil {
		return err
	}
	return nil
}

func LeaveNotifyHandler(X *xgb.Conn) error {
	cookie := xproto.UngrabKeyboardChecked(X, xproto.TimeCurrentTime)
	if err := cookie.Check(); err != nil {
		return err
	}
	return nil
}

func UnmapNotifyHandler(quitfunc func()) {
	fmt.Println("unmap notify event")
	fmt.Println("connection interrupted")
	quitfunc()
}
