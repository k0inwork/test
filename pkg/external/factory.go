package external

// Factory pattern for resolving the correct mock strategy

func NewMessageBroker(mode, endpoint string) MessageBroker {
	switch mode {
	case "real":
		// Here we would return the actual RabbitMQ implementation
		// e.g. return &RealRabbitMQBroker{URL: endpoint}
		// For now, since we only have mock bounds in Phase 1:
		return &MockBroker{}
	case "emulated":
		return &EmulatedBroker{Endpoint: endpoint}
	case "mock":
		fallthrough
	default:
		return &MockBroker{}
	}
}
