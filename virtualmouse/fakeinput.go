package main

import (
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgb/xtest"
)

type ConnInfo struct {
	Conn  *xgb.Conn
	Wid   xproto.Window
	Setup *xproto.SetupInfo
}

func EstablishConn() (*ConnInfo, error) {
	X, err := xgb.NewConnDisplay(":3")
	if err != nil {
		return nil, err
	}
	setup := xproto.Setup(X)
	screen := setup.DefaultScreen(X)

	xtesterr := xtest.Init(X)
	if xtesterr != nil {
		return nil, xtesterr
	}
	return &ConnInfo{X, screen.Root, setup}, nil
}

func (c *ConnInfo) Fakeinput(x int16, y int16) error {
	a := xtest.FakeInputChecked(c.Conn, 6, 0, 0, c.Wid, x, y, 0)
	err := a.Check()
	if err != nil {
		return err
	}
	return nil
}

func (c *ConnInfo) GetPointer() (int16, int16, error) {
	a := xproto.QueryPointer(c.Conn, c.Wid)
	p, err := a.Reply()
	if err != nil {
		return -1, -1, err
	}
	return p.WinX, p.WinY, nil
}
