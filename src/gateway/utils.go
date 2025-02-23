package gateway

func Error(err error) *Message {
	return &Message{
		Type: "ERROR",
		Body: map[string]interface{}{
			"message": err.Error(),
		},
	}
}
