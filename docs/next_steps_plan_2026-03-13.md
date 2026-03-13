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
    *   **Deprecate** the standalone `gws` Go microservice entirely.
    *   Migrate gateway attributes (state, region) from the `models.Gw` structure directly into `models.Node` within the `product` service.
    *   **Drop** the `Session` model entirely as it is unused legacy code.

**3. `inventory` Microservice (Hardware Management & Topology)**
*   **Goal:** Allow physical hardware modifications and track port-level connections.
*   **Actions:**
    *   Implement POST/DELETE for switches (`/data/switch_create`, `/data/switch/<pk>/delete`).
    *   Extend `models.Switch` to handle topology/connections (`switch_customer_connection_create`).
    *   Implement IPMI/PDU control logic (Power On, Cold Reset) by sending RabbitMQ payloads to the hardware controllers.

## Medium Priority (Services, Configuration, and Async Feedback)

**4. `keyservice` Microservice (Connectivity Orders)**
*   **Goal:** Fully implement the lifecycle (creation and deletion) of a "key service", fully integrating with the `route` external module.
*   **Actions:**
    *   Implement POST/DELETE handlers for `keyservice`. Note that these records are not simply database entries; they represent active routing commands.
    *   **Creation:** Wire the creation of a `KeyService` record to automatically dispatch an asynchronous routing request to the `route` external module (`send_routing_request` logic).
    *   **Deletion:** Wire the deletion (or archiving) of a `KeyService` record to automatically dispatch a delete API command to the `route` external module (`send_delete_request` logic) before updating the database status.

**5. `task` Microservice (Async Updates)**
*   **Goal:** Replicate the live progress bar functionality of legacy Celery tasks.
*   **Actions:**
    *   Convert the `task` service to support Server-Sent Events (SSE) or WebSockets on the `/tasks` endpoint, streaming live status updates for background processes.

**6. `identity` Microservice (LDAP User Management & Onboarding)**
*   **Goal:** Port the user management mutations, the GUI workflow for new user application/LDAP provisioning (`RegistrationApplication`), and provide an Identity Provider (IdP) module if required by downstream services.
*   **Actions:**
    *   **User Mutations:** Implement API endpoints for the three supported mutations: Change Password, Edit User Data, and Block User. Wire these directly into the external LDAP directory to manipulate the data.
    *   **Onboarding Workflow:** Implement REST/GraphQL APIs in `identity` to allow users to apply for access, admins to review applications and generate email verification codes, and a final signup endpoint to create the user profile in LDAP.
    *   **CAS Identity Provider (IdP):** Replicate the legacy `django_mama_cas` functionality by implementing a CAS (or OAuth2/OIDC) server within the `identity` service. This is necessary if external applications or reverse proxies (like Apache with `mod_auth_cas`) still need to authenticate against the PUM system.

**7. Switch Configuration Management**
*   **Goal:** Replicate the TFTP config backup and restore logic.
*   **Actions:**
    *   Implement logic within `inventory` or `external-modules` to trigger `get_config`, `set_config`, and configuration checks on hardware switches.

## Low Priority (Settings and Node Mutations)

**8. `core` Microservice (Global Settings)**
*   **Goal:** Provide a central place for global application settings.
*   **Actions:**
    *   Create a new `core` microservice (or append to `registry`) to serve `core/settings` and `core/readyness` endpoints, migrating away from `system.yaml` if runtime mutation is needed.

## Deprecated / Deferred Scope
*   **Product Node Mutations:** Explicitly creating, editing, deleting, starting, or stopping `Node`s is deprecated. Nodes are strictly imported and managed via external synchronization.
*   **Access Rights (RBAC):** Legacy `accounts/access/` roles and views are explicitly deferred. Basic groups are sufficient for the current sprint.
*   **History Sessions:** `/gws/historysession/` and `Session` logic remains unneeded in the new architecture.
*   **Data Services:** `/services/listdataservice/` remains unneeded in the new architecture.
