package bstore

import (
	"encoding/binary"
	"time"

	"github.com/boltdb/bolt"
)

var boltdb *bolt.DB

func LoadBolt(file string) (*bolt.DB, error) {
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return db, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("history"))
		if b == nil {
			_, err := tx.CreateBucket([]byte("history"))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return db, nil
}

type History struct {
	Date time.Time `json:"date"`
	Loss byte      `json:"loss"`
}

// GetHistory get histroy from boltdb
func GetHistory(id int) ([]History, error) {
	var bh []byte
	boltdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("history"))
		bi := make([]byte, 8)
		binary.BigEndian.PutUint64(bi, uint64(id))
		bh = b.Get(bi)
		return nil
	})
	return Unmarshal(bh), nil
}

func PutHistory(id int, hs []History) error {
	boltdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("history"))
		bi := make([]byte, 8)
		binary.BigEndian.PutUint64(bi, uint64(id))
		err := b.Put([]byte(bi), Marshal(hs))
		if err != nil {
			return err
		}
		return nil
	})
	return nil
}

func Unmarshal(b []byte) (hist []History) {
	for i := 0; i < len(b); i++ {
		if len(b) < i+9 {
			break
		}
		ts := int64(binary.LittleEndian.Uint64(b[i : i+8]))
		i += 8
		loss := b[i]
		var h History
		h.Date = time.Unix(int64(ts), 0)
		h.Loss = loss
		hist = append(hist, h)
	}
	return
}

func Marshal(hist []History) (b []byte) {
	for i := range hist {
		bi := make([]byte, 8)
		binary.LittleEndian.PutUint64(bi, uint64(hist[i].Date.Unix()))
		for x := range bi {
			b = append(b, bi[x])
		}
		b = append(b, hist[i].Loss)
	}
	return
}

func init() {
	boltdb, _ = LoadBolt("db.bolt")
}
