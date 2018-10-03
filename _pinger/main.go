package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/prgra/monik/opinger"
)

func main() {
	p, err := opinger.New()
	if err != nil {
		panic(err)
	}
	st, err := p.Ping("8.8.8.8", 10)
	if err != nil {
		fmt.Println(err)
	}
	spew.Dump(st)
	// time.Sleep(10 * time.Second)
}
