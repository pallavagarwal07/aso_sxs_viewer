package main

import (
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgb/xtest"
)

func fakeinput (X *xgb.Conn, wid xproto.Window, x int16, y int16) {
	xtest.FakeInput(X, 6, 0, 0, wid, x, y, 0)
}