package messaging

type PublisherMessage struct {
	Exchange   string
	RoutingKey string
	Body       string
}
