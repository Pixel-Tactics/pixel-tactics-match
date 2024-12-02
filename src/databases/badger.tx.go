package databases

import (
	"encoding/json"

	"github.com/dgraph-io/badger/v4"
)

type BadgerTx interface {
	Commit() error
	Discard()
	BadgerQuery
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

func (tx *BadgerTxImpl) Delete(key string) error {
	return tx.Txn.Delete([]byte(key))
}
