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
    *   **Authentication Flow:** `/accounts/login/`, `/accounts/logout/`. *Note: The new system relies on Apptron/Hanko for auth.*
    *   **User Management & Mutations:** The original codebase supports three core user mutations: `/accounts/edit/<username>/` (Change User Data), `/accounts/changepassword/` (Change Password), and `/accounts/block/<username>/` (Block User). **Crucially, all of these actions simply perform direct updates on the backend LDAP directory.**
    *   **Registration Application Workflow:** The endpoints under `/accounts/register_application/` represent a complete GUI workflow for new user onboarding.
        1. A user applies for access.
        2. An admin approves the application and assigns an LDAP group (`RegistrationApplicationUpdate`).
        3. The system sends an email with a signup link and verification code.
        4. The user completes signup, which provisions the user account and password directly into LDAP (`RegistrationApplicationSignup`).
    *   **Access Rules (Deprecated for now):** Everything under `accounts/access/` (views, rights, create, delete) is explicitly **not needed yet**. Basic users/groups are sufficient for now.
    *   **Miscellaneous:** `/accounts/activity_list_export/`, `/accounts/support/`, `/accounts/monitoring/`, `/accounts/ssh_session_search`.
*   **Required Action:** Expand `identity` to support the 3 specific user mutations (change password, edit data, block user) by mapping them as RPC/External-Module calls directly to LDAP. The Registration Workflow must be analyzed—if Apptron/Hanko handles auth, it must integrate with these LDAP provisioning steps, or the Go `identity` service must natively replicate the approval workflow.

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

## 4. `gws` App (Gateways & Sessions)
**Original Responsibility:** Internal tracking of gateways and communication sessions (linking two gateways/nodes to track test bytes and subnets).
*   **Status in Go Architecture:** The `gws` service currently exists as a standalone microservice exposing `/gateways` and `/sessions`.
*   **Analysis & Required Action:** In the original legacy codebase, `Gw` objects were **purely internal** entities simply tracking a name, address, state, and region. They do not require exposure as a standalone microservice or a separate entity. The concept of a `Gw` is functionally equivalent to a `Node` in the new system.
    *   **Action:** The standalone `gws` microservice should be **deprecated**. `Gw` attributes and logic must be merged directly into the `models.Node` structure within the `product` microservice.
    *   **Session Management:** The `Session` model is unused legacy code and is **completely unneeded**. All connectivity logic and tracking is actually handled via Key Services.

## 5. `services` App (Connectivity Orders)
**Original Responsibility:** High-level abstractions for key services and data transmission orders.
*   **Ported Features (Implemented in Go `keyservice` service):**
    *   `/services/listkeyservice/` -> `keyservice/keyservices`
    *   `/services/listdataservice/` -> *Deprecated / Dropped*
*   **Missing Features:**
    *   **Key Service Management:** `/services/createkeyservice/`, `/services/keyservice/<pk>/` (view), `/services/keyservice/<pk>/delete/`.
*   **Required Action:** Expand `keyservice` to handle creation and deletion. However, it's critical to note that `KeyService` records are not manually manipulated as static records:
    *   **Creation:** When a `KeyService` is created, it must automatically dispatch an async API command to the `route` external module to configure the physical routing logic (`send_routing_request` in legacy code).
    *   **Deletion/Removal:** When a `KeyService` is deleted or marked as archived, it must automatically dispatch a delete API command (`send_delete_request`) to the `route` external module to tear down the connectivity.

## 6. `products` App (Nodes & Monitoring)
**Original Responsibility:** Virtual/physical node management and Zabbix problem aggregation.
*   **Ported Features (Implemented in Go `product` and `external-data` services):**
    *   `/products/products/` -> `product/nodes`
    *   `/products/products/<pk>/` -> `product/nodes/<pk>`
    *   `/products/nodes-problems/` & `/products/zabbix_problems/` -> Unified into `external-data/problems`.
*   **Missing Features:**
    *   **Node Management & Control (Deprecated):** `/products/create/`, `/products/<pk>/edit/`, `/products/<pk>/delete/`, `/products/<pk>/stop/`, `/products/<pk>/start/`, `/products/spdu_request/`. *Nodes are strictly imported via external data synchronization and are never directly created, modified, started, or stopped manually through the GUI.*
    *   **Console/Terminal:** `/modules/console` (Legacy terminal). *Note: The new `terminal` Go microservice is intended to handle SSH/console access.*
*   **Required Action:** Ensure the new `terminal` service fully replaces the `/modules/console` functionality. Explicit node mutations (create/edit/delete/start/stop) are explicitly **deprecated** and do not need to be ported.

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
1. **Lack of Mutation Support:** Most Go microservices currently only implement `GET` (read-only) operations. `POST`, `PUT`, and `DELETE` handlers must be implemented across `identity`, `inventory`, `network`, and `keyservice`. (Note: `product` node mutations and `gws` are deprecated).
2. **Missing External Controller Integration:** Hardware actions (rebooting switches, starting VMs, allocating IPs) are not yet wired up. The `external-modules` proxy needs to fully implement the RabbitMQ RPC payload generation for DHCP, DNS, IPMI, and PDU control.
3. **Missing Domains:** The `core` (global settings) domain is completely absent from the Go architecture.
