package abills

import (
	"database/sql"
	"encoding/binary"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/prgra/monik/pinger"

	//mysql
	_ "github.com/go-sql-driver/mysql"
)

var ping chan Nas
var nases nasdb

func init() {
	ping = make(chan Nas, 1)
	nases.data = make(map[int]Nas)
	for i := 0; i < 9600; i++ {
		go worker(i)
	}
	go periodic()
	go dper()
}
func dper() {
	for {
		time.Sleep(10 * time.Second)
		log.Println("in chan", len(ping))
	}
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

// PingStat :: stat of ping struct
type PingStat struct {
	Cnt  int
	Recv int
	Min  time.Duration
	Max  time.Duration
	Mid  time.Duration
	Sum  time.Duration
}

func ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

// SortedNases :: s
type SortedNases []Nas

func (s SortedNases) Len() int      { return len(s) }
func (s SortedNases) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SortedNases) Less(i, j int) bool {
	return (s[i].LossPerc > s[j].LossPerc) ||
		(s[i].LossPerc == s[j].LossPerc && s[i].Street+s[i].Build < s[j].Street+s[j].Build) ||
		(s[i].LossPerc == s[j].LossPerc && s[i].Street+s[i].Build == s[j].Street+s[j].Build && ip2int(s[i].IP.IP) < ip2int(s[j].IP.IP))
}

// NasGrep :: search in nases
func NasGrep(nases []Nas, s string) []Nas {
	var n []Nas
	for _, v := range nases {
		s = strings.ToLower(s)
		addr := strings.ToLower(v.Street + " " + v.Build)
		if strings.Index(addr, s) >= 0 ||
			strings.Index(v.IP.String(), s) >= 0 ||
			strings.Index(v.MAC.String(), s) >= 0 {
			n = append(n, v)
		}
	}
	return n
}

// GetNases load nases from DB
func GetNases() error {
	rows, err := db.Query(`SELECT n.id, n.ip, n.mac, n.nas_type, n.name, d.name, s.name, b.number, n.descr 
		FROM nas n 
		LEFT JOIN builds b on b.id = n.location_id
		LEFT JOIN streets s on s.id = b.street_id
		LEFT JOIN districts d on d.id = s.district_id 
		WHERE n.disable=0`)
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
func (h *Nas) Ping(c int) PingStat {
	// log.Printf("start ping: %v", h)
	var res PingStat
	for x := 0; x < c; x++ {
		stat, err := pinger.Ping(h.IP.String(), time.Second*5)
		if err != nil {
			log.Printf("Ping error: %v\n", err)
		}
		// log.Printf("stat: %v", stat)
		res.Cnt++

		if !stat.End.IsZero() {
			res.Recv++
			ti := stat.End.Sub(stat.End)
			res.Mid += ti
			if ti > res.Max {
				res.Max = ti
			}
			if ti < res.Min || res.Min == 0 {
				res.Min = ti
			}
		}

	}
	if res.Recv > 0 {
		mi := int64(res.Sum/time.Millisecond) / int64(res.Recv)
		res.Mid = time.Duration(mi) * time.Millisecond
		// fmt.Println("OK ", h.IP, res, suc, 100-100*suc/c)
		h.LossPerc = byte(100 - 100*res.Recv/c)
	} else {
		h.LossPerc = 100
		// log.Println("loss", h.IP)
	}
	nases.Push(*h)
	return res
}

func periodic() {
	log.Println("start periodic")
	time.Sleep(10 * time.Second)
	for {
		log.Println("periodic loop start")
		keys := nases.GetKeys()
		log.Println("keys", len(keys))
		for _, k := range keys {
			ping <- nases.Get(k)
		}
		time.Sleep(60 * time.Second)
	}
}

func worker(id int) {
	for {
		n, ok := <-ping
		if !ok {
			return
		}
		// log.Println("ping", n)
		stat := n.Ping(10)
		if stat.Recv > 0 {
			// log.Println(n.IP, id, stat)
		}
	}
}

//GetOffline return all offline devices
func GetOffline() []Nas {

	var off []Nas
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
		if v.LossPerc == 100 {
			off = append(off, v)
		}
	}
	nases.mut.RUnlock()
	return off
}

//GetAllNases return all offline devices
func GetAllNases() []Nas {
	var all []Nas
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
		all = append(all, v)
	}
	nases.mut.RUnlock()
	return all
}