package main

import (
	"fmt"
	"testing"
)

func TestFakeInput(t *testing.T) {
	a, err := EstablishConn()
	if err != nil {
		t.Errorf("Encountered error %s by EstablishConn()", err)
	}
	tables := []struct {
		x int16
		y int16
	}{
		{10, 10},
		{250, 250},
		{10, 30},
	}

	for _, table := range tables {

		testname := fmt.Sprintf("%d , %d", table.x, table.y)

		t.Run(testname, func(t *testing.T) {

			errInput := a.Fakeinput(table.x, table.y)
			if errInput != nil {
				t.Errorf("Encountered error %s by Fakeinput()", errInput)
			}
			p, q, errPointer := a.GetPointer()
			if errPointer != nil {
				t.Errorf("Encountered error %s by GetPointer()", errPointer)
			}

			if table.x != p || table.y != q {
				t.Errorf("Fakeinput(%v, %v) moves cursor to (%v , %v) , want (%v , %v)", table.x,
					table.y, p, q, table.x, table.y)
			}
		})
	}

}
