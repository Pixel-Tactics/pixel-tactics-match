package config

import "os"

var BadgerPath string

func Setup() {
	BadgerPath = os.Getenv("BADGER_PATH")
}
