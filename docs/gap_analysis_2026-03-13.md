# PUM Final Gap Analysis: Python (Django) vs. Go Microservices

This document provides a comprehensive, final gap analysis based on a direct mapping between the raw legacy Python/Django codebase (from the `demo` repository) and the newly implemented Go microservices.

It lists the legacy endpoints across all Django apps and identifies whether they have been correctly ported to the new Go architecture, highlighting the remaining missing features and required actions.

## 1. `accounts` App (Identity & Access Control)
**Original Responsibility:** User/group management, authentication, access rules, and activity auditing.
*   **Ported Features (Implemented in Go `identity` service):**
    *   `/accounts/list/` -> `identity/users`
    *   `/accounts/currentuser` -> `identity/users/current`
    *   `/accounts/group/list` -> `identity/groups`
    *   `/accounts/activitylist/` -> `identity/activitylist`
*   **Missing Features:**
    *   **Authentication Flow:** `/accounts/login/`, `/accounts/logout/`, `/accounts/changepassword/`, `/accounts/create/` (Registration). *Note: The new system relies on Apptron/Hanko for auth, but user profile syncing might be needed.*
    *   **User Management Actions:** `/accounts/edit/<username>/`, `/accounts/block/<username>/`, `/accounts/unblock/<username>/`, `/accounts/delete/<username>/`.
    *   **Access Rules (Deprecated for now):** Everything under `accounts/access/` (views, rights, create, delete) is explicitly **not needed yet**. Basic users/groups are sufficient for now.
    *   **Miscellaneous:** `/accounts/activity_list_export/`, `/accounts/register_application/` (OAuth/Tokens?), `/accounts/support/`, `/accounts/monitoring/`, `/accounts/ssh_session_search`.
*   **Required Action:** Expand `identity` to support user mutations (edit, block, delete). Access Rules (RBAC) mapping can be deferred as it is not currently required.

## 2. `data` App (Inventory & Hardware)
**Original Responsibility:** Management of switches, PDUs, IPMI nodes, configuration management, and physical connections.
*   **Ported Features (Implemented in Go `inventory` service):**
    *   `/data/switch/` -> `inventory/switches`
    *   `/data/pdu/list` -> `inventory/pdus`
    *   `/data/ipmi/list` -> `inventory/ipmi`
*   **Missing Features:**
    *   **CRUD Operations:** `/data/switch_create/`, `/data/switch/<pk>/delete/`, `/data/new` (Add new equipment).
    *   **Hardware Control (via RabbitMQ/External):**
        *   Switch state: `/data/<pk>/start`, `/data/<pk>/stop`, `/data/<pk>/check`.
        *   IPMI commands: `/data/ipmi/<ip>/poweron`, `poweroff`, `coldreset`, `warmreset`.
        *   PDU commands: `/data/pdu/<ip>/<rele>/restart`, `/data/pdu/<ip>/status`.
    *   **Configuration Management (TFTP/Configs):** Everything under `/data/switch/<ip>/` (status, add_device, get_config, list_config, set_config, check_config).
    *   **Topology/Connections:** `/data/<pk>/switch_customer_connection_create/`, `/data/switch_customer_connections_by_eq`.
*   **Required Action:** Expand the `inventory` microservice to support mutations (create/delete). The critical gap is **Hardware Control** and **Configuration Management**, which require expanding `external-modules` (RabbitMQ integration) to actually trigger state changes on physical/emulated devices. Topology tracking must also be added to the Go database models.

## 3. `network` App (IPAM)
**Original Responsibility:** DHCP, DNS, and Subnet allocation.
*   **Ported Features (Implemented in Go `network` service):**
    *   `/network/subnetwork/` -> `network/subnets`
    *   `/network/dhcp/` -> `network/dhcp`
    *   `/network/dns/` -> `network/dns`
*   **Missing Features:**
    *   **Subnet Management:** `/network/subnetwork/create/`, `/network/subnetwork/delete/`.
    *   **DHCP Management:** `/network/dhcp/create/`, `/network/dhcp/edit/`, `/network/dhcp/delete/`, `/network/dhcp/lease/`.
    *   **DNS Management:** `/network/dns/create/`, `/network/dns/edit/`, `/network/dns/delete/`.
*   **Required Action:** The Go `network` service currently only supports GET requests (reading state). It must be expanded to support POST/PUT/DELETE mutations, which likely require sending specific RPC payloads to backend DHCP/DNS controllers (via RabbitMQ/`external-modules`).

## 4. `gws` App (Gateways & Tunnels)
**Original Responsibility:** Overlay networks, gateways, and VXLAN session orchestration.
*   **Ported Features (Implemented in Go `gws` service):**
    *   `/gws/gws/` -> `gws/gateways`
    *   `/gws/historysession/` -> *Deprecated / Dropped*
    *   `/gws/createsession/` -> Partially covered by POST to `gws/sessions` (logic might be incomplete).
*   **Missing Features:**
    *   **Gateway Mutations:** `/gws/create/`, `/gws/<pk>/edit/`, `/gws/<pk>/delete/`.
    *   **Session Management:** `/gws/session/<pk>/`, `/gws/session/<pk>/delete/`, `/gws/sessiontest/<pk>/`, `/gws/ownsession`.
*   **Required Action:** Expand `gws` service to support full CRUD for Gateways and Sessions. Session creation/deletion must orchestrate the actual VXLAN tunnel configurations on network devices.

## 5. `services` App (Connectivity Orders)
**Original Responsibility:** High-level abstractions for key services and data transmission orders.
*   **Ported Features (Implemented in Go `keyservice` service):**
    *   `/services/listkeyservice/` -> `keyservice/keyservices`
    *   `/services/listdataservice/` -> *Deprecated / Dropped*
*   **Missing Features:**
    *   **Key Service Management:** `/services/createkeyservice/`, `/services/keyservice/<pk>/` (view), `/services/keyservice/<pk>/delete/`.
*   **Required Action:** Expand `keyservice` to handle creation and deletion. Creating a key service requires complex internal orchestration: triggering network switch configuration via the `gws` or `network` services.

## 6. `products` App (Nodes & Monitoring)
**Original Responsibility:** Virtual/physical node management and Zabbix problem aggregation.
*   **Ported Features (Implemented in Go `product` and `external-data` services):**
    *   `/products/products/` -> `product/nodes`
    *   `/products/products/<pk>/` -> `product/nodes/<pk>`
    *   `/products/nodes-problems/` & `/products/zabbix_problems/` -> Unified into `external-data/problems`.
*   **Missing Features:**
    *   **Node Management:** `/products/create/`, `/products/<pk>/edit/`, `/products/<pk>/delete/`.
    *   **Node Control:** `/products/<pk>/stop/`, `/products/<pk>/start/`, `/products/spdu_request/`.
    *   **Console/Terminal:** `/modules/console` (Legacy terminal). *Note: The new `terminal` Go microservice is intended to handle SSH/console access.*
*   **Required Action:** Implement mutations for nodes. Wire up the start/stop controls to the `inventory` service or hardware controllers. Ensure the new `terminal` service fully replaces the `/modules/console` functionality.

## 7. `tasks` App (Asynchronous Jobs)
**Original Responsibility:** Tracking background tasks (Celery).
*   **Ported Features (Implemented in Go `task` service):**
    *   `/tasks/viewtasks/` -> `task/tasks`
*   **Missing Features:**
    *   **Live Updates:** The legacy system streamed real-time progress via WebSockets (Django Channels). The Go `task` service currently returns a static JSON list.
*   **Required Action:** The Go `task` service must implement Server-Sent Events (SSE) or WebSockets to stream asynchronous task state updates to the new frontend.

## 8. `core` App (System Settings)
**Original Responsibility:** Global configurations and readiness checks.
*   **Missing Features (Entirely Un-ported):**
    *   `/core/settings`, `/core/add_setting`, `/core/change_setting/<pk>`, `/core/delete_setting/<pk>`.
    *   `/core/internals`, `/core/externals`, `/core/externals/turnon`, `/core/readyness`.
*   **Required Action:** The new architecture relies heavily on `system.yaml` for external dependencies, but runtime global settings might still be necessary. A `core` or `settings` microservice (or extending `registry`) is needed to expose `/core/settings` and readiness health checks.

---
**Summary of Major Technical Debts in Go:**
1. **Lack of Mutation Support:** Most Go microservices currently only implement `GET` (read-only) operations. `POST`, `PUT`, and `DELETE` handlers must be implemented across `identity`, `inventory`, `network`, `gws`, `keyservice`, and `product`.
2. **Missing External Controller Integration:** Hardware actions (rebooting switches, starting VMs, allocating IPs) are not yet wired up. The `external-modules` proxy needs to fully implement the RabbitMQ RPC payload generation for DHCP, DNS, IPMI, and PDU control.
3. **Missing Domains:** The `core` (global settings) domain is completely absent from the Go architecture.
