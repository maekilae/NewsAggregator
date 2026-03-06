package db

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

type DB struct {
	db *badger.DB
}

func InitDB() (*DB, error) {
	db, err := badger.Open(badger.DefaultOptions("./data/db/"))
	if err != nil {
		return &DB{}, err
	}

	return &DB{db: db}, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Insert(key, value []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (db *DB) Get(key []byte) ([]byte, error) {
	var value []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			value = make([]byte, len(val))
			copy(value, val)
			return nil
		})
	})
	return value, err
}

func (db *DB) Exists(key []byte) bool {
	err := db.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		return err
	})
	if err != nil {
		return false
	}
	return true
}

func (db *DB) GetByPrefix(prefix []byte) ([][]byte, error) {
	var values [][]byte
	err := db.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			_ = item.Key()
			err := item.Value(func(v []byte) error {
				value := make([]byte, len(v))
				copy(value, v)
				values = append(values, value)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return values, err
}

func (db *DB) Delete(key []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (db *DB) Iterate() error {
	return db.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				fmt.Printf("key=%s, value=%s\n", k, v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}
