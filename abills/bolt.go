package abills

import (
	"encoding/binary"
	"time"

	"github.com/boltdb/bolt"
)

var boltdb *bolt.DB

func loadBolt(file string) (*bolt.DB, error) {
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
	return bhUnmarshal(bh), nil
}

func bhUnmarshal(b []byte) (hist []History) {
	for i := 0; i < len(b); i++ {
		if len(b) < i+9 {
			break
		}
		ts, _ := binary.Uvarint(b[i : i+8])
		i += 9
		loss := b[i]
		var h History
		h.Date = time.Unix(int64(ts), 0)
		h.Loss = loss
		hist = append(hist, h)
	}
	return
}
