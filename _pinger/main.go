package main

import (
	"github.com/prgra/monik/opinger"
)

func main() {
	p := opinger.New()
	err := p.Listen()
	if err != nil {
		panic(err)
	}
	_, err = p.Ping("8.8.8.8", 10)
	if err != nil {
		panic(err)
	}

	// time.Sleep(10 * time.Second)
}
