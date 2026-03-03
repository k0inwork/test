# External Module Mocking Strategy

To support rapid development, local testing without Docker, and the transition from Python to Go, the new microservices architecture uses a three-tier mocking strategy for all external dependencies (LDAP, GLPI, Zabbix, RabbitMQ, BPKGW, and Hardware Controllers).

This is configured globally in the root `system.yaml` file under `external_modules`.

## Available Modes

Every external module dependency can be configured into one of three states:

### 1. `mock` (Internal Go Stub)
- **Description:** No actual network calls or external connections are made. The Go microservice uses an internal interface implementation that immediately returns a predefined success/failure response.
- **Use Case:** Local unit testing, UI development, and running the system quickly without any external dependencies running.
- **Implementation:** Interfaces in `pkg/external/` provide a `*MockClient` struct. When `mode: mock` is read from config, the dependency injection layer wires this mock. Example: `task` service simulating a RabbitMQ RPC response internally instead of connecting to a broker.

### 2. `emulated` (External Dummy Service)
- **Description:** The Go microservice acts as if it is connecting to the real service and sends actual network traffic over HTTP/TCP/AMQP to the configured `endpoint`. However, the endpoint points to a lightweight, standalone dummy server (often a small Python or Go script) running locally.
- **Use Case:** Integration testing, verifying network timeouts, testing data serialization/deserialization across boundaries.
- **Implementation:** The service wires a `*RealClient` but points the connection string to `localhost` ports where lightweight dummy containers or scripts answer requests with static JSON payloads.

### 3. `real` (Production/Staging Service)
- **Description:** The Go microservice connects to the actual, live external system.
- **Use Case:** Staging environments, Production deployments.
- **Implementation:** The service wires a `*RealClient` and connects to the production network endpoint, requiring valid credentials and network access.

## Supported Modules

The following modules must respect this configuration across all Go microservices:

*   **`ldap`** (Identity Service): Directory authentication.
*   **`glpi`** (Inventory/Product Services): CMDB synchronization.
*   **`zabbix`** (External-Data Service): Polling problems and alerts.
*   **`bpkgw`** (GWS Service): Orchestrating VXLAN tunnels and routing.
*   **`hardware_controllers`** (External-Modules Service): Proxying IPMI, PDU, and switch config commands.
*   **`rabbitmq`** (Task/Network Services): Message broker for asynchronous actions and RPC.

## Example Configuration (`system.yaml`)

```yaml
external_modules:
  zabbix:
    mode: "mock" # Can be "mock", "emulated", or "real"
    endpoint: "http://localhost:8080/zabbix"
  bpkgw:
    mode: "emulated"
    endpoint: "http://localhost:5000/api/v1"
```
