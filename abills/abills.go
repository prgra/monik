package abills

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"time"

	//mysql
	_ "github.com/go-sql-driver/mysql"
	fastping "github.com/tatsushid/go-fastping"
)

var ping chan Nas
var nases nasdb

func init() {
	ping = make(chan Nas, 10000)
	nases.data = make(map[int]Nas)
	for i := 0; i < 256; i++ {
		go worker()
	}
	go periodic()
}

var db *sql.DB

//New init abills mysql conect and object
func New(url string) error {
	var err error
	log.Println("connect to db")
	db, err = sql.Open("mysql", url)
	GetNases()
	return err
}

type nasdb struct {
	data map[int]Nas
	mut  sync.RWMutex
}

func (n *nasdb) Get(key int) Nas {
	n.mut.RLock()
	defer n.mut.RUnlock()
	return n.data[key]
}

//GetKeys all keys of db
func (n *nasdb) GetKeys() []int {
	var keys []int
	n.mut.RLock()
	for k := range n.data {
		keys = append(keys, k)
	}
	n.mut.RUnlock()
	return keys
}

func (n *nasdb) Push(v Nas) {
	n.mut.Lock()
	n.data[v.ID] = v
	n.mut.Unlock()
}

//Nas abills nas
type Nas struct {
	ID          int
	IP          net.IPAddr
	MAC         net.HardwareAddr
	LossPerc    byte
	PingDate    time.Time
	NasType     string
	Name        string
	Distrinct   string
	Street      string
	Build       string
	Description string
}

// GetNases load nases from DB
func GetNases() error {
	rows, err := db.Query(`SELECT n.id, n.ip, n.mac, n.nas_type, n.name, d.name, s.name, b.number, n.descr 
		FROM nas n 
		LEFT JOIN builds b on b.id = n.location_id
		LEFT JOIN streets s on s.id = b.street_id
		LEFT JOIN districts d on d.id = s.district_id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var n Nas
		var ip, mac, nasType, dname, nname, sname, bnum, descr sql.NullString
		rows.Scan(&n.ID, &ip, &mac, &nasType, &nname, &dname, &sname, &bnum, &descr)
		n.IP.IP = net.ParseIP(ip.String)
		n.MAC, _ = net.ParseMAC(mac.String)
		n.NasType = nasType.String
		n.Name = nname.String
		n.Street = sname.String
		n.Build = bnum.String
		n.Description = descr.String
		nases.Push(n)
	}

	return nil
}

// Ping send to ping chan
func (h *Nas) Ping(c int) {
	p := fastping.NewPinger()
	p.AddIPAddr(&h.IP)
	var suc int
	var min time.Duration
	var max time.Duration
	var sum time.Duration
	p.OnRecv = func(addr *net.IPAddr, elapsed time.Duration) {
		// log.Println("recv")
		suc++

		if min == 0 {
			min = elapsed
		}
		if elapsed < min {
			min = elapsed
		}
		if elapsed > max {
			max = elapsed
		}
		sum += elapsed
		// log.Println("elapsed", elapsed)
	}
	// p.OnIdle = func() {
	// 	log.Println("finish")
	// }
	// log.Println("run")

	for x := 0; x < c; x++ {
		err := p.Run()
		if err != nil {
			fmt.Println(err)
		}
	}
	if suc > 0 {
		mi := int64(sum/time.Millisecond) / int64(suc)
		mid := time.Duration(mi) * time.Millisecond
		fmt.Println("OK ", h.IP, min, mid, max, suc, 100-100*suc/c)
		h.LossPerc = byte(100 - 100*suc/c)
	} else {
		h.LossPerc = 100
		// log.Println("loss", h.IP)
	}
	nases.Push(*h)
}

func periodic() {
	for {
		keys := nases.GetKeys()
		for _, k := range keys {
			ping <- nases.Get(k)
		}
		time.Sleep(20 * time.Second)
	}
}
func worker() {
	for {
		n, ok := <-ping
		if !ok {
			return
		}
		// log.Println("ping", n)
		n.Ping(5)
	}
}

//GetOffline return all offline devices
func GetOffline() ([]Nas, []Nas) {

	var off, on []Nas
	nases.mut.RLock()
	var keys []int
	for k := range nases.data {
		keys = append(keys, k)
	}
	nases.mut.RUnlock()
	sort.Ints(keys)
	nases.mut.RLock()
	for _, k := range keys {
		v := nases.data[k]
		if v.LossPerc > 0 {
			off = append(off, v)
		} else {
			on = append(on, v)
		}
	}
	nases.mut.RUnlock()
	return off, on
}
