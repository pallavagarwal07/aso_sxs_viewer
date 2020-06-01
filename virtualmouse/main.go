package main

import (
	"fmt"
)

func main() {

	c := establishConn()
    var x , y int16
    for {
		fmt.Println("enter x and y coordinates of cursor:")
		fmt.Scan(&x, &y)
		c.fakeinput(x , y)
	}
}