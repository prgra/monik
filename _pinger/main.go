package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/prgra/oping"
)

func main() {
	p, err := oping.New(oping.Conf{Workers: 20000})
	if err != nil {
		panic(err)
	}
	time.Sleep(5 * time.Second)
	h := make(map[string]bool)
	var hm sync.Mutex
	go deb(h, &hm)
	var wg sync.WaitGroup
	for y := 0; y < 255; y++ {
		for x := 0; x < 255; x++ {
			fmt.Printf("ping 10.128.%d.%d\n", y, x)
			hm.Lock()
			h[fmt.Sprintf("10.128.%d.%d", y, x)] = true
			hm.Unlock()
			go func(x int, y int, wg *sync.WaitGroup, p *oping.Pinger) {
				wg.Add(1)
				st, err := p.Ping(fmt.Sprintf("10.128.%d.%d", y, x), 10)
				if err != nil {
					fmt.Println(err)
				}
				suc := 0
				hm.Lock()
				delete(h, fmt.Sprintf("10.128.%d.%d", y, x))
				hm.Unlock()
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
	log.Println("wait...")
	wg.Wait()
	p.Close()
	log.Println("????????????????????????")
}

func deb(h map[string]bool, hm *sync.Mutex) {
	for {
		hm.Lock()
		log.Printf("len of q=%d", len(h))
		hm.Unlock()
		time.Sleep(1 * time.Second)
	}
}
