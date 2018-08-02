package pinger

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

type pstat struct {
	mut  sync.Mutex
	ping int
	wait int
}

var st pstat
var c *icmp.PacketConn
var dbmut sync.RWMutex
var pingdb map[string]chan Stat

var cn chan net.IP

type Stat struct {
	IP    net.IP
	Start time.Time
	End   time.Time
	R     icmp.Message
}

func dper() {
	for {
		time.Sleep(1 * time.Second)
		st.mut.Lock()
		log.Println("waint", st.wait)
		st.mut.Unlock()
	}
}

func Ping(ip string, t time.Duration) (Stat, error) {

	var res Stat
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}

	wb, err := wm.Marshal(nil)
	if err != nil {
		return res, err
	}
	if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(ip)}); err != nil {
		return res, err
	}
	res.IP = net.ParseIP(ip)
	res.Start = time.Now()
	dbmut.Lock()
	if _, ok := pingdb[ip]; ok {
		pingdb[ip] = make(chan Stat)
	}
	dbmut.Unlock()
	select {
	case st := <-pingdb[ip]:
		fmt.Printf("%q\n", st.R)
		dbmut.Lock()
		delete(pingdb, ip)
		dbmut.Unlock()

		return res, nil
	case <-time.After(t):
		return res, fmt.Errorf("timeout :%v", ip)

	}
}

func init() {
	var err error
	pingdb = make(map[string]chan Stat)
	c, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Fatalf("listen err, %s", err)
	}
	go mainloop()
	go dper()
}

// Close listen connection
func Close() error {
	return c.Close()
}

// func timeOutCheck() {

// }

func mainloop() {
	rb := make([]byte, 1500)
	for {
		n, peer, err := c.ReadFrom(rb)
		if err != nil {
			log.Fatal(err)
		}
		rm, err := icmp.ParseMessage(1, rb[:n])
		if err != nil {
			log.Fatal(err)
		}
		dbmut.Lock()
		_, ok := pingdb[peer.String()]
		dbmut.Unlock()

		if ok {
			log.Println(peer.String())
			select {
			case v := <-pingdb[peer.String()]:
				v.End = time.Now()
				v.R = *rm
				pingdb[peer.String()] <- v
			case <-time.After(time.Millisecond * 10):
				log.Println("nonono")
			}
		}

	}
}
