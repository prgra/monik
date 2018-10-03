package opinger

import (
	"fmt"
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

// TODO дописать

// Pinger type for ...
type Pinger struct {
	dbmut sync.RWMutex
	db    map[ping]chan Stat
	c     *icmp.PacketConn
}

// New return new pinger
func New() (*Pinger, error) {
	var p Pinger
	p.db = make(map[ping]chan Stat)
	err := p.Listen()
	return &p, err
}

// Listen start listen icmp proccess in background
func (p *Pinger) Listen() (err error) {
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
			body := rm.Body.(*icmp.Echo)

			p.dbmut.Lock()
			_, ok := p.db[ping{IP: peer.String(), Seq: body.Seq}]
			log.Println(p.db)

			if ok {
				log.Println(peer.String())
				var st Stat
				st.Recv = true
				st.RecvTime = time.Now()
				p.db[ping{IP: peer.String(), Seq: body.Seq}] <- st
			}
			p.dbmut.Unlock()
			// log.Println("no me", spew.Sdump(body.Seq))
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

func (p *Pinger) Ping(ip string, c int) (ra []Stat, err error) {

	var res Stat

	var vg sync.WaitGroup
	for x := 1; x <= c; x++ {
		log.Println("raskolbas", x)
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
		if _, err := p.c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(ip)}); err != nil {
			// return ra, err
		}
		res.SendTime = time.Now()
		p.dbmut.Lock()
		p.db[ping{IP: ip, Seq: x}] = make(chan Stat)
		p.dbmut.Unlock()
		log.Println("send", x)

		log.Println("cach", x)
		go func(x int, vg *sync.WaitGroup) {
			vg.Add(1)
			select {
			case st := <-p.db[ping{IP: ip, Seq: x}]:
				fmt.Println("recv ", ping{IP: ip, Seq: x}, st)
				p.dbmut.Lock()
				delete(p.db, ping{IP: ip, Seq: x})
				p.dbmut.Unlock()
				st.SendTime = res.SendTime
				ra = append(ra, res)
				vg.Done()
			case <-time.After(time.Second * 10):
				log.Println("timeout", x)
				res.Recv = false
				p.dbmut.Lock()
				delete(p.db, ping{IP: ip, Seq: x})
				p.dbmut.Unlock()
				ra = append(ra, res)
				vg.Done()
			}
		}(x, &vg)

		time.Sleep(1000 * time.Millisecond)
	}
	log.Println("waiting")
	vg.Wait()
	return ra, nil
}
