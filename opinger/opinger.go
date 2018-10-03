package opinger

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"

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
func New() *Pinger {
	var p Pinger
	p.db = make(map[ping]chan Stat)
	return &p
}

// Listen start listen icmp proccess in background
func (p *Pinger) Listen() (err error) {
	p.c, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}

	rb := make([]byte, 1500)
	go func() {
		log.Println("listen 0.0.0.0")
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
				p.db[ping{IP: peer.String(), Seq: body.Seq}] <- Stat{Recv: true}
			}
			p.dbmut.Unlock()
			log.Println("no me", spew.Sdump(body.Seq))
		}
	}()
	return nil
}

func (p *Pinger) Ping(ip string, c int) (res Stat, err error) {
	seq := 1
	t := time.Second * 5
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: seq,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}

	wb, err := wm.Marshal(nil)
	if err != nil {
		return res, err
	}
	if _, err := p.c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(ip)}); err != nil {
		return res, err
	}
	res.SendTime = time.Now()
	p.dbmut.Lock()

	if _, ok := p.db[ping{IP: ip, Seq: seq}]; !ok {
		p.db[ping{IP: ip, Seq: seq}] = make(chan Stat)
	}
	p.dbmut.Unlock()
	select {
	case st := <-p.db[ping{IP: ip, Seq: seq}]:
		fmt.Println("recv ", ping{IP: ip, Seq: seq}, st)
		p.dbmut.Lock()
		delete(p.db, ping{IP: ip, Seq: seq})
		p.dbmut.Unlock()

		return res, nil
	case <-time.After(t):
		return res, fmt.Errorf("timeout :%v", ip)

	}
}
