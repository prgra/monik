package abills

import (
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
	Loss int       `json:"loss"`
}

// func GetHistory
func GetHistory(id int) ([]History, error) {
	var h []History
}
