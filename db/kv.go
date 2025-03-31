package db

import (
	"encoding/json"
	"errors"

	bolt "go.etcd.io/bbolt"
)

type CData struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	HomeURI      string `json:"home_uri"`
	RedirectURI  string `json:"redirect_uri"`
}

func InitKvStore(file string) (*bolt.DB, error) {
	return bolt.Open(file, 0644, nil)
}

func AddClient(db *bolt.DB, data *CData) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("oauth_clients"))
		if err != nil {
			return err
		}

		if res, err := json.Marshal(data); err != nil {
			return err
		} else {
			return bucket.Put([]byte(data.ClientID), res)
		}
	})
}

func AddCode(db *bolt.DB, clientID, code string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("oauth_code"))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(clientID), []byte(code))
	})
}

func GetCode(db *bolt.DB, clientID string) (string, error) {
	var code string
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("oauth_code"))
		if bucket == nil {
			return errors.New("bucket \"oauth_code\" doesn't exist")
		}

		code = string(bucket.Get([]byte(clientID)))
		return nil
	})
	return code, err
}

func DelCode(db *bolt.DB, clientID string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("oauth_code"))
		if bucket == nil {
			return errors.New("bucket \"oauth_code\" doesn't exist")
		}
		return bucket.Delete([]byte(clientID))
	})
}

func GetClient(db *bolt.DB, clientID string) (*CData, error) {
	var cData CData
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("oauth_clients"))
		if bucket == nil {
			return errors.New("bucket \"oauth_clients\" doesn't exist")
		}

		res := bucket.Get([]byte(clientID))
		if res == nil {
			return errors.New("invalid client_id")
		}

		if err := json.Unmarshal(res, &cData); err != nil {
			return err
		}
		return nil
	})
	return &cData, err
}

func DeleteClient(db *bolt.DB, clientID string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("oauth_clients"))
		if bucket == nil {
			return errors.New("bucket \"oauth_clients\" doesn't exist")
		}
		return bucket.Delete([]byte(clientID))
	})
}
