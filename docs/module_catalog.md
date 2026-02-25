# External Module Catalog

This document provides a detailed breakdown of the external system APIs that the PUM system interacts with via the `Module.getModule` pattern. These modules represent specialized controllers and external systems developed by different engineering teams.

## 1. Network & Infrastructure Management

### 1.1. `ip-module` (DNS/DHCP Controller)
- **Implementation:** `DNSDHCPModule` in `modules/dnsdhcp.py`.
- **Inferred Purpose:** Manages network identity services (IP allocation, DNS records).
- **API Surface:**
  - `load(command)`: Retrieves current configuration or state.
  - `send(data)`: Dispatches new network configurations.
- **Data Format:** Exchanges data primarily in **YAML** format, wrapped in RabbitMQ messages.
- **Interaction Context:** Called by the `network` app to provision or update network settings for managed devices.

### 1.2. `switch-module` (Switch Configuration Controller)
- **Implementation:** `SwitchConfigModule` in `modules/equipment.py`.
- **Inferred Purpose:** Deep configuration and state monitoring of network switches.
- **API Surface:** `status`, `get_from_device`, `add_new_switch`, `add_new_config`, `set`, `check`.
- **Dynamic Mapping:** Uses `__getattr__` to map Python method calls to space-separated RabbitMQ commands (e.g., `add_new_switch` becomes `"add new switch"`).
- **Interaction Context:** Used for automated switch provisioning and verifying configuration consistency across the network fabric.

### 1.3. `route-module` (Routing & Session Controller)
- **Implementation:** `RouteModule` in `modules/route.py`.
- **Inferred Purpose:** Orchestrates L2/L3 sessions and path selection across the gateway fabric.
- **API Surface:**
  - `create(port, bandwidth, priority, uid, session)`: Establishes a new connection.
  - `remove(port, id, session)`: Tears down an existing connection.
- **Interaction Context:** Heavily used by `services/tasks.py` during session establishment and teardown. It returns JSON responses containing established IP addresses and connection IDs.

## 2. Hardware Control

### 2.1. `pdu-module` (Power Distribution Unit Controller)
- **Implementation:** `PDUModule` in `modules/equipment.py`.
- **Inferred Purpose:** Controls physical power supply to racks and individual devices.
- **API Surface:**
  - `statusRelayAll(ip)`: Gets the power state of all outlets on a PDU.
  - `inverseRelay(ip, number, interval)`: Toggles the state of a specific power outlet (reboot).
- **Security:** Requires hardcoded credentials (`PDU_DEVICE_USER/PASSWORD`) passed in every request.

### 2.2. `ipmi-module` (Server Out-of-Band Controller)
- **Implementation:** `IPMIModule` in `modules/equipment.py`.
- **Inferred Purpose:** Remote server management (Baseboard Management Controller interaction).
- **API Surface:** `status`, `power_on`, `soft_power_off`, `hard_reset`, `ipmi_cold_reset`.
- **Interaction Context:** Provides "last resort" control for unresponsive servers, allowing the system to perform hard reboots or monitor hardware health independent of the OS.

## 3. External System Integrations

### 3.1. `asux-module` (ASU-X Integration)
- **Implementation:** `ASUXModule` in `modules/asu_iks.py`.
- **Inferred Purpose:** Integration with a higher-level or parallel monitoring/control system (ASU-X).
- **API Surface:** `request(skip_wait=True)`.
- **Interaction Context:** Used to synchronize state or "turn on" external system tracking. It supports asynchronous requests where the system doesn't wait for an immediate response.

### 3.2. `ready-module` (Health & Readiness Factor)
- **Implementation:** `ReadyModule` in `modules/ready.py`.
- **Inferred Purpose:** Aggregates health metrics into a single "Readiness" factor for the managed environment.
- **API Surface:** `getdata()`.
- **Interaction Context:** Called by the dashboard views to display the overall "health" score of the entire managed facility.

## 4. Interaction Summary Table

| Module | Protocol | Format | Typical Consumer |
|--------|----------|--------|------------------|
| `ip-module` | RabbitMQ | YAML | `network/actions.py` |
| `switch-module` | RabbitMQ | JSON | `data/tasks.py` |
| `route-module` | RabbitMQ | JSON | `services/tasks.py` |
| `pdu-module` | RabbitMQ | JSON | `data/actions.py` |
| `ipmi-module` | RabbitMQ | JSON | `data/tasks.py` |
| `asux-module` | RabbitMQ | JSON | `core/views.py` |
| `ready-module` | RabbitMQ | JSON | `core/views.py` |
