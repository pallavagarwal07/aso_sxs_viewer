package main

import (
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgb/xtest"
)

func main() {

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

	var x , y int16

	for {

		fmt.Println("enter x and y cooridnates of cursor:")
		fmt.Scan(&x, &y)

		fakeinput(X, screen.Root, x , y)
		}
	}