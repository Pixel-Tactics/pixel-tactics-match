package databases

import (
	"encoding/json"

	"github.com/dgraph-io/badger/v4"
	"pixeltactics.com/match/src/config"
)

type Badger interface {
	NewReadTransaction() BadgerTx
	NewReadWriteTransaction() BadgerTx
	Close()
	BadgerQuery
	BadgerBatchQuery
}

type BadgerSetParams struct {
	Key   string
	Value interface{}
}

type BadgerDeleteParams struct {
	Key string
}

type BadgerImpl struct {
	badger *badger.DB
}

func (db *BadgerImpl) NewReadTransaction() BadgerTx {
	return &BadgerTxImpl{
		Txn: db.badger.NewTransaction(false),
	}
}

func (db *BadgerImpl) NewReadWriteTransaction() BadgerTx {
	return &BadgerTxImpl{
		Txn: db.badger.NewTransaction(true),
	}
}

func (db *BadgerImpl) Get(key string, dest interface{}) error {
	return db.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, dest)
		})
	})
}

func (db *BadgerImpl) Set(key string, value interface{}) error {
	return db.badger.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return txn.Set([]byte(key), data)
	})
}

func (db *BadgerImpl) BatchSet(params []*BadgerSetParams) error {
	return db.badger.Update(func(txn *badger.Txn) error {
		for _, param := range params {
			data, err := json.Marshal(param.Value)
			if err != nil {
				return err
			}
			err = txn.Set([]byte(param.Key), data)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *BadgerImpl) Delete(key string) error {
	return db.badger.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (db *BadgerImpl) BatchDelete(params []*BadgerDeleteParams) error {
	return db.badger.Update(func(txn *badger.Txn) error {
		for _, param := range params {
			err := txn.Delete([]byte(param.Key))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *BadgerImpl) Close() {
	db.badger.Close()
}

func NewBadgerImpl() *BadgerImpl {
	db, err := badger.Open(badger.DefaultOptions(config.BadgerPath))
	if err != nil {
		panic("invalid badger path")
	}
	return &BadgerImpl{
		badger: db,
	}
}
