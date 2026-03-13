# PUM Transition Next Steps Plan - 2026-03-13

Based on the latest gap analysis (`gap_analysis_2026-03-13.md`), the following actionable steps are required to reach feature parity between the Go microservices and the legacy Python/Django PUM application.

These tasks are prioritized by domain criticality.

## High Priority (Core Missing Domains & Mutations)

**1. `network` Microservice (IPAM: DHCP, DNS, Subnets)**
*   **Goal:** Expand the existing `network` service to allow editing state, not just reading it.
*   **Actions:**
    *   Implement POST/PUT/DELETE handlers for `/network/subnets`.
    *   Implement POST/PUT/DELETE handlers for `/network/dhcp` and `/network/dns`.
    *   Integrate these mutations with the `external-modules` proxy/RabbitMQ logic to dispatch physical configuration commands to the backend `ip-module`.

**2. `product` Microservice (Node Refactoring & Gateways)**
*   **Goal:** Merge the internal concept of `Gw` (Gateways) into the `Node` model, as Gateways were purely internal entities in the legacy architecture.
*   **Actions:**
    *   **Deprecate** the standalone `gws` Go microservice.
    *   Migrate gateway attributes (state, region) from the `models.Gw` structure into `models.Node` within the `product` service.
    *   Migrate `Session` tracking logic (linking two nodes to track `tx_bytes` and assigned subnets) into the `product` or `network` microservice as an internal associative model, rather than requiring standalone public `gws/sessions` endpoints.

**3. `inventory` Microservice (Hardware Management & Topology)**
*   **Goal:** Allow physical hardware modifications and track port-level connections.
*   **Actions:**
    *   Implement POST/DELETE for switches (`/data/switch_create`, `/data/switch/<pk>/delete`).
    *   Extend `models.Switch` to handle topology/connections (`switch_customer_connection_create`).
    *   Implement IPMI/PDU control logic (Power On, Cold Reset) by sending RabbitMQ payloads to the hardware controllers.

## Medium Priority (Services, Configuration, and Async Feedback)

**4. `keyservice` Microservice (Connectivity Orders)**
*   **Goal:** Fully implement the creation of a "key service".
*   **Actions:**
    *   Implement POST/DELETE handlers for `keyservice`.
    *   Wire the service creation to link target nodes and provision IP/subnets via the `network` service.

**5. `task` Microservice (Async Updates)**
*   **Goal:** Replicate the live progress bar functionality of legacy Celery tasks.
*   **Actions:**
    *   Convert the `task` service to support Server-Sent Events (SSE) or WebSockets on the `/tasks` endpoint, streaming live status updates for background processes.

**6. Switch Configuration Management**
*   **Goal:** Replicate the TFTP config backup and restore logic.
*   **Actions:**
    *   Implement logic within `inventory` or `external-modules` to trigger `get_config`, `set_config`, and configuration checks on hardware switches.

## Low Priority (Settings and Node Mutations)

**7. `core` Microservice (Global Settings)**
*   **Goal:** Provide a central place for global application settings.
*   **Actions:**
    *   Create a new `core` microservice (or append to `registry`) to serve `core/settings` and `core/readyness` endpoints, migrating away from `system.yaml` if runtime mutation is needed.

**8. `product` Node Actions**
*   **Goal:** Control node state.
*   **Actions:**
    *   Implement Start/Stop logic for products/nodes (`/products/<pk>/start`, `/products/<pk>/stop`).
    *   Implement CRUD logic for `/products/create` and `/products/<pk>/edit`.

## Deprecated / Deferred Scope
*   **Access Rights (RBAC):** Legacy `accounts/access/` roles and views are explicitly deferred. Basic groups are sufficient for the current sprint.
*   **History Sessions:** `/gws/historysession/` remains unneeded in the new architecture.
*   **Data Services:** `/services/listdataservice/` remains unneeded in the new architecture.
