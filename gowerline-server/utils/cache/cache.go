package cache

import (
	"encoding/json"
	"fmt"
	"sync"

	bolt "go.etcd.io/bbolt"
)

type SimpleCache struct {
	mutex      *sync.Mutex
	bucketName string
	db         *bolt.DB
}

func NewSimpleCache(bucketName string, db *bolt.DB) (*SimpleCache, error) {
	cache := &SimpleCache{
		bucketName: bucketName,
		mutex:      &sync.Mutex{},
		db:         db,
	}

	return cache, cache.Init()
}

func (c *SimpleCache) Init() error {
	tx, err := c.db.Begin(true)
	if err != nil {
		return nil
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.CreateBucketIfNotExists([]byte(c.bucketName))
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (c *SimpleCache) Put(key string, value interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	tx, err := c.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	bucket, err := tx.CreateBucketIfNotExists([]byte(c.bucketName))
	if err != nil {
		return err
	}

	b, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	err = bucket.Put([]byte(key), b)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (c *SimpleCache) Get(key string, value interface{}) (bool, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	tx, err := c.db.Begin(false)
	if err != nil {
		return false, err
	}
	defer tx.Rollback() //nolint:errcheck

	bucket := tx.Bucket([]byte(c.bucketName))
	b := bucket.Get([]byte(key))
	if b == nil {
		return false, fmt.Errorf("bucket %s not found", c.bucketName)
	}

	err = json.Unmarshal(b, &value)
	if err != nil {
		return false, err
	}

	return value == nil, nil
}
