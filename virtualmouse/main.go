package main

import "fmt"

func main() {
	c, err := EstablishConn()
	if err != nil {
		fmt.Println(err)
		return
	}
	var x, y int16
	for {
		fmt.Println("enter x and y coordinates of cursor:")
		fmt.Scan(&x, &y)
		err = c.Fakeinput(x, y)
		if err != nil {
			fmt.Println(err)
			return
		}

	}
}
