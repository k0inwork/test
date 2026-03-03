package external

// NewMessageBroker acts as a factory, resolving the requested configuration mode
// ("mock", "emulated", or "real") into the appropriate implementation of the MessageBroker interface.
// This ensures microservices do not hardcode dependency strategies and can switch seamlessly
// from local development to production.
func NewMessageBroker(mode, endpoint, realEndpoint string) MessageBroker {
	switch mode {
	case "real":
		// In a production environment, this would instantiate and return a struct
		// managing an actual AMQP connection using the production URL.
		// e.g., `&RealRabbitMQBroker{URL: realEndpoint}`
		// Because the current system is in Phase 1 (mock/stub mode), it falls back to a MockBroker.
		return &MockBroker{}
	case "emulated":
		// Returns a broker configured to communicate with a standalone dummy server.
		return &EmulatedBroker{Endpoint: endpoint}
	case "mock":
		fallthrough
	default:
		// Returns an in-memory stub that performs no networking.
		return &MockBroker{}
	}
}
