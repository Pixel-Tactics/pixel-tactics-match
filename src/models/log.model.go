package models

type SessionLog struct {
	Type      string
	SessionId string
	Data      map[string]interface{}
}
