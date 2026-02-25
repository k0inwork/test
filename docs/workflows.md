# Critical System Workflows

This document breaks down the step-by-step logic for key system processes.

## 1. Multi-Master Coordination (Failover)

The system ensures that only one ARM is "Master" for a given region, while others are "Passive".

1. **Trigger:** An ARM starts up or a `check_arm_master` periodic task runs.
2. **Detection:** The ARM checks `GlobalSettings` for its status. If status is unknown, it broadcasts a `get_master` command over the `amqp.fanout` exchange.
3. **Response:** The current Master receives the `get_master` request and broadcasts its identity and reason for being Master.
4. **Transition:** If a transition is needed (e.g., via manual command), a `change_master` message is broadcast.
5. **Acknowledgment:** Receiving ARMs update their local `GlobalSettings` to "PASSIVE" (read-only mode), inform connected Web UI clients via WebSockets, and send a `change_master_ack`.
6. **Persistence:** The Master ARM tracks acknowledgments in Redis (`REDIS_PROCESSED_REQUEST`) to ensure the transition is complete.

## 2. Gateway Session (VXLAN) Creation

1. **Trigger:** User requests a connection between two gateways in the UI.
2. **Model Creation:** A `Session` object is created in the Django database.
3. **API Call:** The Django backend sends a POST request to the `bpkgw` REST API on the target gateways.
4. **Low-Level Config:** The `bpkgw` service (Flask) executes `ip link add vxlan...` to create the tunnel.
5. **State Update:** If successful, the `Session` status is updated, and traffic monitoring begins.

## 3. Product Synchronization (GLPI)

1. **Periodic Trigger:** `reload_periodic_sync` runs every few minutes.
2. **Data Fetch:** The task calls `glpi_get_datacenters_2`, which queries the GLPI API.
3. **Mapping:** The system maps GLPI entities to the `Product` model using the `sync()` method.
4. **Conflict Resolution:** `sync_search()` is used to find existing records by matching complex naming patterns.
5. **Update:** New products are created, and existing ones are updated with current status and geo-location.

## 4. ASU-X External Routing

1. **Message Arrival:** An external system (`ASU-X`) sends a message to the `asuiks_channels` RabbitMQ queue.
2. **Special Worker:** The special Celery worker receives the message and triggers `asu_iks_routing`.
3. **Processing:**
   - The system validates endpoint switches and ports.
   - It creates a `KeyService` record.
   - It starts a background thread (`wait_for_it`) to monitor the status of the routing task.
4. **Response:** Once the routing task completes (tracked via `TaskRecord`), the system reports the result back to `ASU-X` via a reply-to queue in RabbitMQ.
