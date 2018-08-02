package main

import (
	"monik/pinger"
	"strconv"
	"time"
)

func main() {
	for x := 1; x < 255; x++ {

		go pinger.Ping("192.168.1."+strconv.Itoa(x), time.Second*5)
	}
	time.Sleep(100 * time.Second)
}
