package types

type Request struct {
	Message *Message
}

type Response struct {
	SendToClient      func(message *Message)
	NotifyClient      func(message *Message)
	NotifyOtherClient func(playerId string, message *Message)
	RegisterPlayer    func(playerId string)
	// SendToClient func(data map[string]interface{}, action MessageAction)
	// SendToClientsInSession func() error
}

type Interaction struct {
	Request  *Request
	Response *Response
}
