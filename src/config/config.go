package config

import "os"

var BadgerPath string
var UserServiceUrl string

func Setup() {
	BadgerPath = os.Getenv("BADGER_PATH")
	UserServiceUrl = os.Getenv("USER_MICROSERVICE_URL")
}
