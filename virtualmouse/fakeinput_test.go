package main 

import (
	"fmt"
	"testing"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgb/xtest"
)

func TestFakeInput(t *testing.T) {
	X, err := xgb.NewConnDisplay(":3")
	if err != nil {
		fmt.Println(err)
		return
	}
	
    setup := xproto.Setup(X)
	screen := setup.DefaultScreen(X)

	xtesterr := xtest.Init(X)
	if xtesterr != nil {
		fmt.Printf("Error: %s\n", xtesterr)
	}

	tables := []struct {
		x int16
		y int16
	}{
		{10 , 10},
		{250 , 250},
		{10 , 30},
	}

    for _ , table := range tables {
		
		testname := fmt.Sprintf("%d , %d", table.x, table.y)
		
		t.Run(testname, func(t *testing.T) {

		fakeinput(X, screen.Root, table.x, table.y)
		c := xproto.QueryPointer(X , screen.Root)
		p, ptrqueryerr := c.Reply()
		
		if ptrqueryerr != nil {
		}
		
		if table.x != p.WinX || table.y != p.WinY {
			t.Errorf("didn't work")
		}
	})
	}

}