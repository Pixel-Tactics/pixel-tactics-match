package config

import "os"

var RMQUrl string
var UserServiceUrl string

func Setup() {
	RMQUrl = os.Getenv("RABBITMQ_URL")
	UserServiceUrl = os.Getenv("USER_MICROSERVICE_URL")
}
