package main 

import (
	"fmt"
	"testing"
)

func TestFakeInput(t *testing.T) {
	a := establishConn()
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

		a.fakeinput(table.x, table.y)
		/*c := xproto.QueryPointer(a.conn, a.wid)
		p, ptrqueryerr := c.Reply()
		
		if ptrqueryerr != nil {
		}*/
		p,q := a.getPointer()
		
		if table.x != p || table.y != q {
			t.Errorf("fakeinput(%v, %v) moves cursor to (%v , %v) , want (%v , %v)",table.x, 
			table.y, p, q, table.x, table.y)
		}
	})
	}

}