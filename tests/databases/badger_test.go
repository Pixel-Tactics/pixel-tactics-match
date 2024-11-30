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
