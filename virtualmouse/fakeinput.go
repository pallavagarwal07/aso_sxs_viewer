package main

import (
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgb/xtest"
)

type connInfo struct {
	 conn *xgb.Conn
	 wid xproto.Window
}

func establishConn() connInfo {
	X, err := xgb.NewConnDisplay(":3")
	if err != nil {
		fmt.Println(err)
		/*return statement here should be?*/ 
	}
	setup := xproto.Setup(X)
	screen := setup.DefaultScreen(X)

	xtesterr := xtest.Init(X)
	if xtesterr != nil {
		fmt.Printf("Error: %s\n", xtesterr)
	}
	return connInfo{X, screen.Root}
}

func (c *connInfo) fakeinput (x int16, y int16) {
	xtest.FakeInput(c.conn, 6, 0, 0, c.wid, x, y, 0)
}

func (c *connInfo) getPointer () (x int16, y int16) {
	a := xproto.QueryPointer(c.conn, c.wid)
	p, _ := a.Reply()
    return p.WinX, p.WinY
}