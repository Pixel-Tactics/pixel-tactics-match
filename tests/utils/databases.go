package test_utils

import (
	"log"
	"os"

	"pixeltactics.com/match/src/config"
	"pixeltactics.com/match/src/databases"
)

func NewTestBadger() databases.Badger {
	config.BadgerPath = "./temp/badger-test"
	return databases.NewBadgerImpl()
}

func CloseTestBadger(db databases.Badger) {
	db.Close()
	if config.BadgerPath == "" || config.BadgerPath == "/" {
		panic("invalid path")
	}
	err := os.RemoveAll(config.BadgerPath)
	log.Println("REMOVNIG")
	if err != nil {
		panic(err)
	}
}
