package main

import "fmt"
import "rudp"
import "time"
import "github.com/woodywanghg/gofclog"

func main() {

	fclog.Init(true, true, "rudp.log", 1048576, fclog.LEVEL_DEBUG)

	var obj rudp.ReliableUdp
	err := obj.Init("0.0.0.0", 8008)
	if err != nil {
		fmt.Printf("Init server error! err=%s\n", err.Error())
		return
	}

	for {
		time.Sleep(1000000 * 10000)

	}
}
