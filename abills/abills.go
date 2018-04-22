package abills

import (
	"database/sql"
	"log"
	"net"

	//mysql
	_ "github.com/go-sql-driver/mysql"
)

func init() {

}

var db *sql.DB

//New init abills mysql conect and object
func New(url string) error {
	var err error
	log.Println("connect to db")
	db, err = sql.Open("mysql", url)
	return err
}

//Nas abills nas
type Nas struct {
	ID          int
	IP          net.IP
	MAC         net.HardwareAddr
	NasType     string
	Name        string
	Distrinct   string
	Street      string
	Build       string
	Description string
}

func GetNases() ([]Nas, error) {
	var nases []Nas
	rows, err := db.Query(`SELECT n.id, n.ip, n.mac, n.nas_type, n.name, d.name, s.name, b.number, n.descr 
		FROM nas n 
		LEFT JOIN builds b on b.id = n.location_id
		LEFT JOIN streets s on s.id = b.street_id
		LEFT JOIN districts d on d.id = s.district_id`)
	if err != nil {
		return nases, err
	}
	for rows.Next() {

		var n Nas
		var ip, mac, nasType, dname, nname, sname, bnum, descr sql.NullString
		rows.Scan(&n.ID, &ip, &mac, &nasType, &nname, &dname, &sname, &bnum, &descr)
		n.IP = net.ParseIP(ip.String)
		n.MAC, _ = net.ParseMAC(mac.String)
		n.NasType = nasType.String
		n.Name = nname.String
		n.Street = sname.String
		n.Build = bnum.String
		n.Description = descr.String
		nases = append(nases, n)
	}

	return nases, nil
}
