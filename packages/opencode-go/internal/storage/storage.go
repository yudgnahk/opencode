package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"
)

type Storage struct {
	db *bolt.DB
}

func New(path string) (*Storage, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	// Create buckets
	err = db.Update(func(tx *bolt.Tx) error {
		buckets := []string{"sessions", "messages", "projects", "config"}
		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		db.Close()
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

// Generic get/set methods
func (s *Storage) Get(bucket, key string) ([]byte, error) {
	var data []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		data = b.Get([]byte(key))
		return nil
	})
	return data, err
}

func (s *Storage) Set(bucket, key string, value []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Put([]byte(key), value)
	})
}

func (s *Storage) Delete(bucket, key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Delete([]byte(key))
	})
}

// Helper for JSON marshaling
func (s *Storage) GetJSON(bucket, key string, v interface{}) error {
	data, err := s.Get(bucket, key)
	if err != nil {
		return err
	}
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, v)
}

func (s *Storage) SetJSON(bucket, key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.Set(bucket, key, data)
}

// List all keys in bucket
func (s *Storage) List(bucket string) ([]string, error) {
	var keys []string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.ForEach(func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})
	return keys, err
}
