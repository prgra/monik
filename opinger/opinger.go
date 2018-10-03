package opinger

import (
	"log"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Stat one icmp statistic
type Stat struct {
	SendTime time.Time
	RecvTime time.Time
	Recv     bool
	Size     int
}

type ping struct {
	IP  string
	Seq int
}

type pingParam struct {
	ip    string
	cnt   int
	rchan chan []Stat
}

// TODO дописать

// Pinger type for ...
type Pinger struct {
	dbmut    sync.RWMutex
	db       map[ping]chan Stat
	c        *icmp.PacketConn
	workerwg sync.WaitGroup
	ch       chan pingParam
}

// New return new pinger
func New() (*Pinger, error) {
	var p Pinger
	p.db = make(map[ping]chan Stat)
	p.ch = make(chan pingParam)
	err := p.listen()
	for x := 0; x < 5000; x++ {
		p.workerwg.Add(1)
		go p.pingWorker()

	}
	return &p, err
}

func (p *Pinger) Dump() {
	for {
		p.dbmut.Lock()
		log.Printf("DEBUG::in channel %d, db:%d", len(p.ch), len(p.db))
		p.dbmut.Unlock()
		time.Sleep(time.Second * 1)
	}
}

func (p *Pinger) pingWorker() error {
	for {
		v, ok := <-p.ch
		if !ok {
			break
		}
		st, err := p.rping(v)
		if err != nil {
			return err
		}
		v.rchan <- st
	}
	p.workerwg.Done()
	return nil
}

func (p *Pinger) Close() {
	close(p.ch)
	p.workerwg.Wait()
	// p.c.Close()
}
func (p *Pinger) Ping(ip string, cnt int) ([]Stat, error) {
	var pp pingParam
	pp.rchan = make(chan []Stat)
	pp.ip = ip
	pp.cnt = cnt
	p.ch <- pp
	res := <-pp.rchan
	return res, nil
}

// Listen start listen icmp proccess in background
func (p *Pinger) listen() (err error) {
	p.c, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}

	rb := make([]byte, 1500)
	go func() {
		// log.Println("listen 0.0.0.0")
		for {
			n, peer, err := p.c.ReadFrom(rb)
			if err != nil {
				log.Fatal(err)
			}
			rm, err := icmp.ParseMessage(1, rb[:n])
			if err != nil {
				log.Fatal(err)
			}

			switch body := rm.Body.(type) {
			case *icmp.Echo:

				p.dbmut.Lock()
				ch, ok := p.db[ping{IP: peer.String(), Seq: body.Seq}]
				p.dbmut.Unlock()
				// log.Println(p.db)
				if ok {
					// log.Println(peer.String())
					var st Stat
					st.Recv = true
					st.RecvTime = time.Now()
					ch <- st
				}
				// log.Println("no me", spew.Sdump(body.Seq))

			}
		}
	}()
	return nil
}

func generateData(c int) []byte {
	var date []byte
	for x := 0; x < c; x++ {
		date = append(date, byte(x))
	}
	return date
}

// func (p *Pinger) waitForLim() {
// 	for {
// 		p.dbmut.Lock()
// 		cnt := len(p.db)
// 		p.dbmut.Unlock()
// 		if cnt < p.limit {
// 			log.Println("return from lim")
// 			return
// 		}
// 		time.Sleep(500 * time.Millisecond)
// 		log.Println("limit ", cnt)
// 	}

// }

func (p *Pinger) rping(pp pingParam) (ra []Stat, err error) {
	var res Stat
	var vg sync.WaitGroup
	for x := 0; x < pp.cnt; x++ {
		// log.Println("raskolbas", x)
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID: os.Getpid() & 0xffff, Seq: x,
				Data: []byte(generateData(128)),
			},
		}

		wb, err := wm.Marshal(nil)
		if err != nil {
			// return ra, err
		}
		if _, err := p.c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(pp.ip)}); err != nil {
			// return ra, err
		}
		res.SendTime = time.Now()
		p.dbmut.Lock()
		p.db[ping{IP: pp.ip, Seq: x}] = make(chan Stat)
		p.dbmut.Unlock()
		// log.Println("send", x)

		// log.Println("cach", x)

		go func(x int, wg *sync.WaitGroup, res Stat) {
			wg.Add(1)
			p.dbmut.Lock()
			ch := p.db[ping{IP: pp.ip, Seq: x}]
			p.dbmut.Unlock()
			select {
			case st := <-ch:
				// fmt.Println("recv ", ping{IP: ip, Seq: x}, st)
				p.dbmut.Lock()
				delete(p.db, ping{IP: pp.ip, Seq: x})
				p.dbmut.Unlock()
				st.SendTime = res.SendTime
				st.Recv = true
				st.RecvTime = time.Now()
				ra = append(ra, st)
				wg.Done()
			case <-time.After(time.Second * 10):
				// log.Println("timeout", x)
				res.Recv = false
				p.dbmut.Lock()
				delete(p.db, ping{IP: pp.ip, Seq: x})
				p.dbmut.Unlock()
				ra = append(ra, res)
				wg.Done()
			}
		}(x, &vg, res)

		time.Sleep(1000 * time.Millisecond)
	}
	// log.Println("waiting")
	vg.Wait()
	return ra, nil
}
