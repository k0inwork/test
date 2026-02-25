# Multi-Master Middleware and Read-Only Enforcement

To maintain data consistency in a distributed ARM environment, the system implements a strict read-only enforcement mechanism on "Slave" (Passive) instances.

## 1. `FilterPOSTOnSlave` Middleware

The `FilterPOSTOnSlave` middleware (in `products_app/middleware/middleware.py`) is the gatekeeper for all data-modifying operations.

- **Operation:**
  1. It identifies the current instance's identity (`SETTINGS_TSUM`) and the cluster's active master (`SETTINGS_CONTROL`) from `GlobalSettings`.
  2. For every incoming `POST` request, it checks if the current instance is the Master.
  3. **If Passive:** It raises a `PermissionDenied` exception for most POST requests, effectively making the entire web application read-only.
- **Exceptions:** Critical paths like `/accounts/login/`, `/accounts/logout/`, and `/core/` (used for state management) are exempted to allow basic system operation and state transitions.

## 2. Dynamic Address Discovery
The middleware also performs passive discovery of access points. It monitors the `Host` header of incoming requests and automatically registers new access addresses in `GlobalSettings (SETTINGS_ADDRESSES)`. This ensures that the Multi-Master coordination system always has an up-to-date list of addresses for all ARM instances in the cluster.

## 3. Startup Initialization (`startup_middleware`)
The `startup_middleware` is a "one-shot" middleware that runs once during system boot. It:
- Initializes the system's Active/Passive state based on the Multi-Master configuration.
- Populates `GlobalSettings` with default values defined in `core.variables`, ensuring the database is always in a consistent, functional state.
