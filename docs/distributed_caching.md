# Distributed Caching System

The PUM system utilizes a highly sophisticated caching layer built on top of Redis, providing more than just simple key-value storage.

## 1. Advanced Caching Decorators (`core/cache/utils.py`)

### 1.1. `@periodic_cache`
This decorator is designed for data that is expensive to fetch but can be updated on a schedule.
- **Background Refresh:** Instead of refreshing on a user request, these caches are updated by Celery workers (`reload_periodic_cache`).
- **State Tracking:** It maintains two versions in Redis: `CURRENT` and `OLD`. This allows the system to perform **Change Detection**.
- **On-Update Hooks:** Developers can provide an `on_update` callback (with `filter` and `result` functions) that is triggered only when the cached data actually changes.

### 1.2. `@redis_cached`
A more traditional expiring cache with several advanced features:
- **Async Awareness:** Native support for decorating `async def` functions.
- **Force Reload:** Supports a `reload=True` parameter to bypass the cache and fetch fresh data.
- **Automated Invalidation:** Provides a `.clear_cache()` method on the decorated function for manual invalidation.

## 2. Distributed Lifecycle

1. **Registration:** When a function is decorated, it automatically registers itself in the `REDIS_PERIODIC_CACHE` list in Redis.
2. **Execution:** The decorator first checks Redis for a valid, non-expired result.
3. **Synchronization:** For `periodic_cache`, a Celery worker iterates through the registered functions, calls them, compares the result with the `OLD` version, and triggers update hooks if necessary.

This architecture ensures that the Web UI remains extremely responsive by serving data from Redis, while heavy lifting (like polling 100+ hardware devices) is handled entirely in the background.
