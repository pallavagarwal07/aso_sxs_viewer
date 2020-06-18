package main

import (
	"fmt"
	"os"
	"os/signal"
)

func main() {
	ackSignal := make(chan os.Signal, 1)
	signal.Notify(ackSignal, os.Interrupt)
	fmt.Println("Hello")

	<-ackSignal
	fmt.Println("World")
}
