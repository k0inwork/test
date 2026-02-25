# Celery and RabbitMQ Analysis

This document provides a deep dive into how the system utilizes Celery and RabbitMQ, focusing on connection management and the "special worker" pattern.

## 1. RabbitMQ Connection Management

To avoid exhausting connections and channels, the system uses a custom singleton pattern for `pika`.

### 1.1. PikaChannelSingleton (`core.pika.py`)
- **Pattern:** Singleton wrapper around `pika.BlockingConnection`.
- **Purpose:**
  - Ensures only one connection is opened per process for each connection type.
  - Automatically handles reconnection if the connection is dropped or closed.
  - Shared across Celery tasks and the Django application.
- **Efficiency:** This is critical for the system's stability, especially when many short-lived tasks need to communicate with RabbitMQ.

## 2. The "Special Worker" Pattern

The system employs two types of Celery workers:

### 2.1. Standard Celery Worker
- **Broker:** Redis.
- **Usage:** Standard asynchronous tasks (e.g., database maintenance, periodic syncs).
- **Configuration:** Defined in `products_app/settings.py` via `CELERY_BROKER_URL`.

### 2.2. RabbitMQ-Bound Worker (Special Worker)
- **Library:** `event_consumer`
- **Mechanism:** This worker uses the `AMQPRetryConsumerStep` in Celery.
- **Functionality:** It listens to specific RabbitMQ queues and maps incoming messages to Python functions decorated with `@message_handler`.
- **Use Case:** Inter-ARM communication (Multi-Master) and integration with external modules like `ASU-X`.

## 3. RabbitMQ Communication Patterns

### 3.1. RPC (Remote Procedure Call)
Implemented in `products/tasks.py` via `RpcClient` and `RpcClientAsync`.
- **Workflow:**
  1. Creates an exclusive, non-durable callback queue.
  2. Publishes a message with a `correlation_id` and `reply_to` set to the callback queue.
  3. Waits (blocking or async) for a response on the callback queue.

### 3.2. Fanout (Broadcasting)
Implemented via `FanoutClient`.
- **Usage:** Used for broadcasting state changes across all ARMs (e.g., `change_master` command).
- **Exchange:** `amqp.fanout`.

## 4. Performance Tuning
- **`CELERY_WORKER_PREFETCH_MULTIPLIER=0`**: Disabled prefetching to ensure fair task distribution among workers, which is important for long-running or IO-bound network tasks.
- **`CELERY_BROKER_POOL_LIMIT=100`**: Allows for a higher number of concurrent connections to the Redis broker.
- **`celery_pool_asyncio`**: The system is configured to use an asyncio-compatible pool for Celery, allowing tasks to be defined as `async def` and use `await`.
