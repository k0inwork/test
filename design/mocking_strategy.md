# External Dependency Mocking Strategy

## 1. Goal
To enable local development and testing of the microservices without requiring access to live external infrastructure such as RabbitMQ, LDAP servers, or hardware-specific modules.

## 2. Abstraction Layers
All interactions with external systems must be performed through interfaces. This allows switching between real implementations and mocks easily.

### 2.1 LDAP (Identity Service)
Instead of direct LDAP calls, the service will use an `AuthProvider` interface.
- **Real implementation**: Uses `go-ldap`.
- **Mock implementation**: Uses a hardcoded map of users or a simple file-based lookup.

### 2.2 RabbitMQ (Hardware/Product Services)
Instead of a live RabbitMQ connection, a `MessageBroker` interface will be used.
- **Real implementation**: Uses `amqp091-go`.
- **Mock implementation**: An in-memory pub/sub or a logger that simply records sent commands.

### 2.3 Hardware Modules (`modules/*`)
The Python `modules/` act as proxies. In Go, we will have a `ModuleRegistry` interface.
- **Mock implementation**: Returns predefined JSON/YAML responses for specific commands (e.g., `statusRelayAll` returns a fixed "ON" state).

## 3. Mock Implementations

### Identity Mocking
```go
type AuthProvider interface {
    Authenticate(username, password string) (bool, error)
    GetUserGroups(username string) ([]string, error)
}
```

### Hardware Mocking
```go
type HardwareProxy interface {
    Call(moduleName string, command string, params map[string]interface{}) (interface{}, error)
}
```
In Phase 1, the `HardwareProxy` implementation will return mock data based on the `moduleName` and `command`.

## 4. Configuration
A configuration flag (e.g., `MOCK_EXTERNAL=true`) will determine which implementation is instantiated at runtime.

## 5. Benefits
- **Deterministic Tests**: No flaky tests due to network issues or external system state.
- **Zero-Install Dev**: New developers can start coding with just Go and SQLite.
- **Hardware Independence**: Develop UI and logic even when target hardware is offline.
