package store

import (
	"errors"

	"go.etcd.io/bbolt"
)

const dataBucket = "data"

type DiskDataStore struct {
	db *bbolt.DB
}

func NewDiskDataStore(filePath string) (*DiskDataStore, error) {
	db, err := bbolt.Open(filePath, 0600, nil)
	if err != nil {
		return nil, err
	}
	if err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(dataBucket))
		return err
	}); err != nil {
		return nil, err
	}
	return &DiskDataStore{db: db}, nil
}

func (d *DiskDataStore) Get(key string) ([]byte, error) {
	var retval []byte
	err := d.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dataBucket))
		if bucket == nil {
			return errors.New("bucket not found")
		}
		value := bucket.Get([]byte(key))
		if value == nil {
			return ErrKeyNotFound
		}
		copy(retval, value)
		return nil
	})
	return retval, err
}

func (d *DiskDataStore) Set(key string, value []byte) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dataBucket))
		if bucket == nil {
			return errors.New("bucket not found")
		}
		return bucket.Put([]byte(key), value)
	})
}

func (d *DiskDataStore) Delete(key string) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dataBucket))
		if bucket == nil {
			return errors.New("bucket not found")
		}
		return bucket.Delete([]byte(key))
	})
}

func (d *DiskDataStore) GenerateId() (uint, error) {
	var id uint
	err := d.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dataBucket))
		if bucket == nil {
			return errors.New("bucket not found")
		}
		nextid, err := bucket.NextSequence()
		if err != nil {
			return err
		}
		id = uint(nextid)
		return nil
	})
	return id, err
}
