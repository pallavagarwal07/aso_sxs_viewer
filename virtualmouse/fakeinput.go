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

func EstablishConn(display string) (*ConnInfo, error) {
	X, err := xgb.NewConnDisplay(display)
	if err != nil {
		return nil, err
	}
	setup := xproto.Setup(X)
	screen := setup.DefaultScreen(X)

	if xtesterr := xtest.Init(X); xtesterr != nil {
		return nil, xtesterr
	}
	return &ConnInfo{X, screen.Root, setup}, nil
}

func (c *ConnInfo) MoveMouse(x int16, y int16) error {
	a := xtest.FakeInputChecked(c.Conn, 6, 0, 0, c.Wid, x, y, 0)
	if err := a.Check(); err != nil {
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
