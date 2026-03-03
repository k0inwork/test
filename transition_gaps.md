# PUM Transition Gap Analysis: Python (Django) to Go Microservices

This document analyzes the current progress of migrating the legacy Python/Django PUM application to the new Go-based microservices architecture, focusing specifically on missing functionality and gaps in the API layer, and how these map to the new static/Apptron GUI goals.

## Current State

The new Go architecture is structured around domain-specific microservices (`identity`, `product`, `inventory`, `network`, `task`, `registry`, `frontend`, `terminal`, `external-modules`, `external-data`, and `compatibility`). The `compatibility` service acts as a proxy bridge, allowing the old Django GUI to function by mapping legacy endpoints to their new Go counterparts.

Currently, several legacy endpoints are correctly mapped and proxied by `compatibility`:
- `/accounts/list/` -> `identity/users`
- `/accounts/currentuser` -> `identity/users/current`
- `/accounts/group/list` -> `identity/groups`
- `/products/products/` -> `product/nodes`
- `/products/products/:pk/` -> `product/nodes/:pk`

However, a significant portion of the original application logic remains un-ported, currently returning "Not Implemented" stubs via the `compatibility` service.

## Identified Gaps (Missing Features)

The following endpoints are stubbed in `compatibility/main.go` and represent functional gaps between the old Python backend and the new Go system. They are grouped by their original Django application domains:

### 1. `data` App (Network Equipment & Inventory)
**Original Responsibility:** Low-level inventory of switches, PDUs, and IPMI interfaces. Tracking physical connections.
- **Stubbed Endpoints:**
  - `/data/switch/` (`switch_list`)
  - `/data/pdu/list` (`pdu_list`)
  - `/data/ipmi/list` (`ipmi_list`)
**Gap:** While the `inventory` Go service exists and has a basic `/switches` endpoint, the `compatibility` layer is not yet routing legacy `/data/switch/` requests to it. Additionally, PDU and IPMI specific tracking and management (previously in `data`) are missing from the Go `inventory` or `external-modules` APIs.

### 2. `network` App (IPAM & DNS/DHCP)
**Original Responsibility:** IP address allocation, DNS, and DHCP management.
- **Stubbed Endpoints:**
  - `/network/dhcp/` (`dhcp_list`)
  - `/network/dns/` (`dns_list`)
  - `/network/subnetwork/` (`subnetwork_list`)
**Gap:** The `network` Go service exists and has a basic `/subnets` endpoint, but DHCP and DNS management capabilities are missing or not exposed. The legacy endpoints are still stubbed.

### 3. `gws` App (Gateways & Sessions)
**Original Responsibility:** Orchestration of network gateways and VXLAN tunnels (sessions).
- **Stubbed Endpoints:**
  - `/gws/gws/` (`gateways`)
  - `/gws/historysession/` (`session_history`)
**Gap:** There is currently no dedicated Go microservice (or clear responsibility within an existing one) for managing network gateways and VXLAN tunnel sessions. This entire domain is missing from the Go architecture.

### 4. `services` App (Connectivity Services)
**Original Responsibility:** High-level abstractions for data and key transmission orders/services.
- **Stubbed Endpoints:**
  - `/services/listkeyservice/` (`key_services`)
  - `/services/listdataservice/` (`data_services`)
**Gap:** Similar to `gws`, the high-level concept of "Key Services" and "Data Services" (connectivity orders) is completely missing from the Go microservices landscape.

### 5. `products` App (Zabbix/Problems)
**Original Responsibility:** Managing products (PDUs/OUs) and integrating with monitoring systems like Zabbix.
- **Stubbed Endpoints:**
  - `/products/nodes-problems/` (`nodes_problems`)
  - `/products/zabbix_problems/` (`zabbix_problems`)
**Gap:** While the basic `/nodes` endpoint is ported to the `product` Go service, the integration for fetching and displaying active problems/alerts (likely from Zabbix or internal tracking) is missing.

### 6. `tasks` App (Operations Tracking)
**Original Responsibility:** Central hub for tracking asynchronous background operations (previously Celery).
- **Stubbed Endpoints:**
  - `/tasks/viewtasks/` (`tasks`)
**Gap:** The `task` Go service exists and has a `/tasks` endpoint, but the `compatibility` layer is not routing legacy requests to it. It remains stubbed.

### 7. `accounts` App (Auditing)
**Original Responsibility:** User management and activity logging.
- **Stubbed Endpoints:**
  - `/accounts/activitylist/` (`activity_list`)
**Gap:** While core Identity (users/groups) is ported, the `ActivityLog` (security auditing and history) functionality is missing.

## GUI Replacement Strategy (Apptron & Static GUI)

The ultimate goal is to replace the old Django server-side rendered GUI with a new, static GUI utilizing the Apptron environment. The `compatibility` layer is a temporary bridge.

**Implications for the New GUI:**
1. **Direct Microservice Communication:** The new static GUI should *not* rely on the `compatibility` service or legacy Django JSON formats (`json_middleware`). It should communicate directly with the new Go microservices (`identity`, `product`, `inventory`, etc.) via standard REST or GraphQL.
2. **Missing Backend Support:** The new GUI cannot fully replace the old one until the missing backend domains (`gws`, `services`, `data` specifics, `network` specifics, `activitylist`) are implemented in Go.
3. **Apptron Integration:** For features like device terminals and advanced hardware control, the new static GUI can leverage Apptron's local-first execution environment (WASM) and its network bridging capabilities, bypassing traditional server-side proxies (like `django_webssh`).

## Transition Recommendations

1. **Un-Stub Existing Services:** Update `services/compatibility/main.go` to proxy `/tasks/viewtasks/` to the existing `task` service (`/tasks`), and `/data/switch/` to the existing `inventory` service (`/switches`).
2. **Implement Missing Domains:** Design and implement new Go microservices (or expand existing ones) for:
   - **Gateways & Tunnels (`gws`)**: Needs a dedicated service for VXLAN orchestration.
   - **Connectivity Services (`services`)**: Needs a service to handle data/key transmission orders.
   - **DHCP/DNS Management**: Expand the `network` service to handle these alongside subnets.
   - **Auditing**: Expand the `identity` or create an `audit` service for the `activitylist`.
3. **Expand External Integrations:** Implement the Zabbix problems logic within the `external-data` or `product` services. Expand `external-modules` to fully support IPMI and PDU specifics beyond the current mocks.
4. **Develop Static GUI Components:** As each domain is ported to Go, build the corresponding view in the new Apptron/static frontend, gradually rendering the legacy Django GUI obsolete.
