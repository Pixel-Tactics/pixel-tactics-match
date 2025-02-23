package databases_test

import (
	"testing"

	"github.com/go-playground/assert/v2"
	"pixeltactics.com/match/src/config"
	"pixeltactics.com/match/src/databases"
)

func TestBadger(t *testing.T) {
	config.BadgerPath = "./temp/badger"

	badger := databases.NewBadgerImpl()
	err := badger.Set("testos", 123)
	assert.Equal(t, err, nil)

	var obj int
	err = badger.Get("testos", &obj)
	assert.Equal(t, err, nil)

	assert.Equal(t, obj, 123)
}

func TestTransaction(t *testing.T) {
	config.BadgerPath = "./temp/badger"

	badger := databases.NewBadgerImpl()

	err := badger.Set("count", 1)
	assert.Equal(t, err, nil)
	err = badger.Set("count", 1)
	assert.Equal(t, err, nil)

	tx1 := badger.NewReadWriteTransaction()
	defer tx1.Discard()
	tx2 := badger.NewReadWriteTransaction()
	defer tx2.Discard()

	var count1 int
	var count2 int
	err = tx1.Get("count", &count1)
	assert.Equal(t, err, nil)
	err = tx2.Get("count", &count2)
	assert.Equal(t, err, nil)
	assert.Equal(t, count1, 1)
	assert.Equal(t, count2, 1)

	err = tx1.Set("count", count1+1)
	assert.Equal(t, err, nil)
	err = tx2.Set("count", count2+1)
	assert.Equal(t, err, nil)

	err = tx2.Commit()
	assert.Equal(t, err, nil)
	err = tx1.Commit()
	assert.Equal(t, err, nil)
}
