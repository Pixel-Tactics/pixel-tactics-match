package databases

import (
	"encoding/json"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type BadgerTx interface {
	Commit() error
	Discard()
	ExtendedQuery
}

type BadgerTxImpl struct {
	Txn *badger.Txn
}

func (tx *BadgerTxImpl) Commit() error {
	return tx.Txn.Commit()
}

func (tx *BadgerTxImpl) Discard() {
	tx.Txn.Discard()
}

func (tx *BadgerTxImpl) Get(key string, dest interface{}) error {
	item, err := tx.Txn.Get([]byte(key))
	if err != nil {
		return err
	}
	return item.Value(func(val []byte) error {
		return json.Unmarshal(val, dest)
	})
}

func (tx *BadgerTxImpl) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return tx.Txn.Set([]byte(key), data)
}

func (tx *BadgerTxImpl) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	e := badger.NewEntry([]byte(key), data).WithTTL(ttl)
	return tx.Txn.SetEntry(e)
}

func (tx *BadgerTxImpl) PrefixScanWithLimit(prefix string, limit int) ([]string, [][]byte, error) {
	it := tx.Txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	arrKeys := make([]string, 0)
	arrValues := make([][]byte, 0)
	prefixByte := []byte(prefix)
	for it.Seek(prefixByte); it.ValidForPrefix(prefixByte); it.Next() {
		if limit > 0 && len(arrValues) == limit {
			break
		}

		item := it.Item()
		arrKeys = append(arrKeys, string(item.Key()))
		val, err := item.ValueCopy(nil)
		if err != nil {
			return nil, nil, err
		}
		arrValues = append(arrValues, val)
	}
	return arrKeys, arrValues, nil
}

func (tx *BadgerTxImpl) PrefixScan(prefix string) ([]string, [][]byte, error) {
	return tx.PrefixScanWithLimit(prefix, -1)
}

func (tx *BadgerTxImpl) Delete(key string) error {
	return tx.Txn.Delete([]byte(key))
}
