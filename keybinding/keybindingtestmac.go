package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"../command"
	"github.com/chromedp/chromedp"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

//just for testing on different enviornment, not the actual test

func main() {
	allocCtx1, cancel := chromedp.NewExecAllocator(context.Background(), execAllocatorOptions1...)
	defer cancel()
	ctx1, cancel := chromedp.NewContext(allocCtx1)
	defer cancel()

	if err := chromedp.Run(ctx1,
		chromedp.Navigate(`https://mail.google.com`),
	); err != nil {
		log.Fatal(err)
	}
	time.Sleep(2 * time.Second)
	BrowserList = append(BrowserList, ctx1)

	X, screenInfo := Newconn()
	CreateInputWindow(0, 0, 1280, 50, X, screenInfo)

}

var execAllocatorOptions1 = []chromedp.ExecAllocatorOption{
	chromedp.NoDefaultBrowserCheck,
	chromedp.Flag("user-data-dir", "/tmp/aso_sxs_viewer/dir1"),
}

func Newconn() (*xgb.Conn, *xproto.ScreenInfo) {
	X, err := xgb.NewConn()

	if err != nil {
		log.Fatalln(err)
	}

	setup := xproto.Setup(X)
	screenInfo := setup.DefaultScreen(X)
	return X, screenInfo
}

func cmdErrorHandler(p *command.ProgramState, err error) error {
	return err
}

//CreateInputWindow creates window to caprure keycodes
func CreateInputWindow(x uint32, y uint32, w uint16, h uint16,
	X *xgb.Conn, screenInfo *xproto.ScreenInfo) {

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

	UpdateMaps(X)
	eventloop(X)
}

func eventloop(X *xgb.Conn) {

	for {
		ev, err := X.WaitForEvent()

		if err != nil {
			// Error Handler for program state
			continue
		}

		if ev == nil {
			fmt.Println("Both event and error are nil: connection interrupted")
		}

		switch event := ev.(type) {
		case xproto.KeyPressEvent:
			KeyPressHandler(X, KeyPressEvent{&event})

		case xproto.MapNotifyEvent:
			UpdateMaps(X)
		}
	}
}
