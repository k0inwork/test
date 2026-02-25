# Redis Usage Analysis

Redis is a multi-purpose component in the system, acting as a message broker, a cache, and a shared state store.

## 1. Custom Storage Abstractions (`core.redisbackend.py`)

The system does not use Redis as a simple key-value store. It implements several abstractions for structured data storage:

### 1.1. RedisStore
- **Role:** An "object-relational" mapper for Redis.
- **Mechanism:** Stores objects as JSON strings. It maintains a set of IDs (`index:<ClassName>:id`) and supports indexing specific fields using Redis Hashes (`index:<ClassName>:<FieldName>`).
- **Use Case:** Storing ephemeral objects like `TaskRecord` that need fast access and expiration but don't require the overhead of a SQL database.

### 1.2. RedisVar
- **Role:** Typed variable storage.
- **Mechanism:** Handles serialization/deserialization of Python types (including `set`) to JSON.
- **Feature:** Includes a `lock()` context manager for distributed locking, ensuring atomic updates to shared variables.

### 1.3. RedisHashTable
- **Role:** Wrapper for Redis Hashes (`HSET`, `HGET`).
- **Mechanism:** Provides a dictionary-like interface to Redis Hashes with automatic tracking of access and modification times.

## 2. Infrastructure Roles

### 2.1. Celery Broker
All standard Celery tasks are queued in Redis. This is separate from the specialized RabbitMQ communication.

### 2.2. Django Channels
Redis serves as the backing store for `CHANNEL_LAYERS`, enabling real-time communication (WebSockets) between the backend and the browser.

### 2.3. Task State & Coordination
- **`REDIS_PERIODIC_CACHE`**: Stores the list and timing of periodic tasks to ensure they are run at the correct intervals across the cluster.
- **`REDIS_PROCESSED_REQUEST`**: Tracks processed Multi-Master requests to prevent duplicate processing.
- **`REDIS_ARM_ID`**: Stores a unique identifier for the current ARM instance, ensuring it can identify its own messages in broadcast exchanges.

## 3. Connection Pooling
`RedisConnection` implements a singleton `ConnectionPool`. This ensures that the application doesn't create a new TCP connection for every Redis operation, significantly improving performance and reducing resource usage.
