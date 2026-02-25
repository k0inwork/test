# Data Synchronization Engine

The `core/cache/sync.py` module implements a robust engine for synchronizing internal Django models with external data sources (like GLPI or specialized hardware controllers).

## 1. Object Mapping and Rename Detection

One of the most complex challenges in distributed systems is tracking objects that might be renamed in the external source. The sync engine solves this using a "Double-Check" strategy:

1. **Check by Index:** It first tries to find the object using its primary identifier (e.g., `name` or `serial_number`) in a Redis Hash (`RedisHashTable`).
2. **Check by Foreign ID:** If not found by name, it searches the Redis Hash for the object's `foreignid` (the ID in the external system).
3. **Rename Handling:** If found by `foreignid`, the system realizes the object has been renamed. It updates the internal Redis mappings and renames the Django model instance to match the new external name.
4. **Creation:** If neither check succeeds, a new Django model instance is created.

## 2. The `model.sync()` Protocol

Models that participate in synchronization must implement a `sync(data)` generator method.
- **Responsibility:** The model's `sync` method is responsible for mapping the external `data` dictionary to its own fields.
- **Generator Pattern:** It uses `yield self` to allow for flexible saving logic and potential side effects (like creating related objects) during the sync process.

## 3. Transactional Safety

The sync engine uses the `reload_periodic_sync` Celery task to orchestrate the process. It collects all objects to be saved and performs database updates using `database_sync_to_async(s.save)()`, ensuring that the database remains the source of truth while leveraging the speed of Redis for object lookup and rename tracking.
