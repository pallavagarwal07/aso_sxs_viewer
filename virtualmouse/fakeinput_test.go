package virtualmouse

import (
	"fmt"
	"testing"
)

func TestMoveMouse(t *testing.T) {
	display := 1
	CreateXephyrWindow(display)
	a, err := EstablishConn(display)
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

			if errInput := a.MoveMouse(table.x, table.y); errInput != nil {
				t.Errorf("Encountered error %s by MoveMouse()", errInput)
			}
			p, q, errPointer := a.GetPointer()
			if errPointer != nil {
				t.Errorf("Encountered error %s by GetPointer()", errPointer)
			}

			if table.x != p || table.y != q {
				t.Errorf("MoveMouse(%v, %v) moves cursor to (%v , %v) , want (%v , %v)", table.x,
					table.y, p, q, table.x, table.y)
			}
		})
	}

}
