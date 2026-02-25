# Module Interaction System

The system employs a registry-based architecture for managing interactions with external services and hardware controllers. This pattern ensures high decoupling between the Django business logic and the underlying communication protocols.

## 1. The Service Locator Pattern (`Module.getModule`)

The core of the system is the `Module` base class in `modules/base.py`, which acts as a registry for all available service modules.

### 1.1. Registration
Modules register themselves during initialization:
```python
Module.registerModule(LazyProxy(lambda: PDUModule(queue)), "pdu-module")
```
Using `LazyProxy` ensures that the module (and its potentially heavy initialization or connection logic) is only instantiated when first accessed.

### 1.2. Retrieval
Any part of the application can obtain a reference to a module without knowing its concrete implementation:
```python
module = Module.getModule("pdu-module")
await module.statusRelayAll(ip="192.168.1.100")
```

## 2. Dynamic Command Mapping

Many modules (e.g., `PDUModule`, `IPMIModule`, `SwitchConfigModule`) use Python's `__getattr__` to dynamically map method calls to RabbitMQ commands.

- **Workflow:**
  1. A method like `statusRelayAll()` is called on a module instance.
  2. Since the method is not explicitly defined, `__getattr__` catches it.
  3. It constructs a payload: `{'command': 'statusRelayAll', 'params': {...}}`.
  4. It dispatches the payload via RabbitMQ and waits for a response.

This design allows the Python code to support new commands from the hardware controllers without requiring code changes in the module definitions.

## 3. Distributed Request Relaying

The system supports a hybrid execution model to prevent long-running RabbitMQ operations from blocking the web application's async loop.

### 3.1. Local vs. Worker Execution
The `RabbitMqModule.call` method checks the `settings.CURRENT_WORKER` flag:
- **If `True` (Inside a Worker):** The module executes the RabbitMQ call directly using `pika`.
- **If `False` (Inside the Web App):** The request is relayed to a background worker via Django Channels.

### 3.2. Async Result Storage
When a request is relayed:
1. The web application generates a `corr_id` and sends the request to the `send-rpc` channel.
2. The `FanoutConsumer` (special worker) receives the request and executes the actual RabbitMQ call.
3. The worker stores the response in a Redis Hash (`variables.RABBITMQ_ASYNC_RESULT_TABLE`) under the `corr_id`.
4. The web application polls or waits for the result in Redis.

## 4. Key Registered Modules

| Alias | Implementation | Responsibility |
|-------|----------------|----------------|
| `pdu-module` | `PDUModule` | Control of Power Distribution Units (relays, state). |
| `ipmi-module` | `IPMIModule` | Out-of-band management for servers (power, resets). |
| `switch-module` | `SwitchConfigModule` | Network switch configuration and status monitoring. |
| `ip-module` | `DNSDHCPModule` | IP address management and configuration. |
| `asux-module` | `ASUXModule` | Integration with the ASU-X external system. |
| `route-module` | `RouteModule` | Network path and session routing management. |
| `ready-module` | `ReadyModule` | Health and readiness Factor monitoring. |

## 5. Security and Access Control
Most modules inherit from `AccessControlledMixin`, allowing them to use the `call_ac` method. This ensures that hardware commands are validated against the user's permissions and the current state of the resource before being dispatched to the network.
