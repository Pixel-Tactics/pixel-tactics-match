package databases

import (
	"github.com/dgraph-io/badger/v4"
	"pixeltactics.com/match/src/config"
)

func SetupBadger() *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(config.BadgerPath))
	if err != nil {
		panic("invalid badger path")
	}
	return db
}
