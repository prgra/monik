package abills

import "github.com/boltdb/bolt"

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
