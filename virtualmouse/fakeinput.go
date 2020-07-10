package virtualmouse

import (
	"fmt"

	"github.com/googleinterns/aso_sxs_viewer/command"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgb/xtest"
)

type ConnInfo struct {
	Conn  *xgb.Conn
	Wid   xproto.Window
	Setup *xproto.SetupInfo
}

func EstablishConn(display int) (*ConnInfo, error) {
	displayString := fmt.Sprintf(":%d", display)
	X, err := xgb.NewConnDisplay(displayString)
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

func CreateXephyrWindow(display int) error {
	displayString := fmt.Sprintf(":%d", display)
	xephyr := command.ExternalCommand{
		Path: "xephyr",
		Arg: []string{
			displayString,
			"-ac",
			"-br",
			"-screen",
			"-reset",
			"-no-host-grab",
		},
	}
	_, err := command.ExecuteProgram(xephyr, cmdErrorHandler)
	if err != nil {
		return err
	}
	return nil
}

func cmdErrorHandler(err error) error {
	if err != nil {
		fmt.Printf("returned error %s, calling force quit", err.Error())
	}
	return err
}
