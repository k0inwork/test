# Advanced Utilities and Monitoring

The `core/utils.py` module contains several sophisticated decorators and engines that power the PUM system's robustness and monitoring capabilities.

## 1. Robust Decorators

### 1.1. `@timeout(max_time)`
A versatile decorator that supports both synchronous and asynchronous functions.
- **Sync Functions:** It uses a separate thread and `queue.Queue` to run the function and joins the thread with the specified timeout.
- **Async Functions:** It leverages `asyncio.wait_for` for native timeout handling.

### 1.2. `@thread_safe`
Ensures that a function is not executed concurrently by multiple processes or threads. It uses the `redis.lock` mechanism described in the Redis analysis. This is essential for preventing race conditions during hardware configuration.

### 1.3. `@retry(retries=3)`
A simple wrapper for automatic retry of failing functions, typically used for network operations that may experience transient failures.

## 2. The Alarm System (`@alarm`)

The `@alarm` decorator is the primary way the system monitors the health of its integrations.

- **Mechanism:**
  1. If a decorated function raises an exception or times out, the decorator catches it.
  2. It triggers the `alarm_raised` Django signal.
  3. A signal receiver in `products_app/signals.py` updates the `GlobalSettings` (`SETTINGS_ALARMS`).
  4. The entire system state (`SETTINGS_STATE`) is changed to **"ALARM"**.
  5. A real-time notification is sent to all connected web clients via Django Channels.
- **Auto-Recovery:** When a subsequent call to the decorated function succeeds, it triggers `alarm_shutdown`, which clears the alarm and restores the system state to "ACTIVE" or "PASSIVE".

## 3. Recursive Serialization Engine (`jsonify`)

The `jsonify` function is a custom, highly recursive serialization engine used throughout the project for converting complex Python and Django objects into JSON-serializable formats.

- **Features:**
  - **Django Model Support:** Automatically serializes Django models, handling foreign keys and many-to-many relationships.
  - **Custom Level Control:** Prevents infinite recursion by allowing a `level` parameter to limit how deep the serialization goes.
  - **Method Serialization (`@json_attr`):** Developers can mark model methods with `@json_attr(level=X)` to include the result of that method in the serialized output.
  - **Type Support:** Built-in handling for UUIDs, Enums, datetime objects (converted to timestamps), and `LazyProxy` objects.
