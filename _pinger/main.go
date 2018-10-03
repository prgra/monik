package main

import (
	"fmt"
	"sync"

	"github.com/prgra/monik/opinger"
)

func main() {
	p, err := opinger.New()
	if err != nil {
		panic(err)
	}
	go p.Dump()
	var wg sync.WaitGroup
	for y := 128; y < 140; y++ {
		for x := 1; x < 255; x++ {
			fmt.Printf("ping 10.128.%d.%d\n", y, x)
			go func(x int, y int, wg *sync.WaitGroup, p *opinger.Pinger) {
				wg.Add(1)
				st, err := p.Ping(fmt.Sprintf("10.128.%d.%d", y, x), 10)
				if err != nil {
					fmt.Println(err)
				}
				suc := 0
				for _, s := range st {
					if s.Recv {
						suc++
					}
				}
				if suc > 0 {
					fmt.Printf("10.128.%d.%d - %d\n", y, x, suc)
				}

				wg.Done()

			}(x, y, &wg, p)
		}
	}
	// time.Sleep(10 * time.Second)
	wg.Wait()
}
