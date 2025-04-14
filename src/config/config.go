package config

import "os"

var UserServiceUrl string

func Setup() {
	UserServiceUrl = os.Getenv("USER_MICROSERVICE_URL")
}
