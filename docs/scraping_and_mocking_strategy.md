# Scraping and Mocking Strategy for the Go Transition

## 1. Overview
The transition of the PUM backend from a legacy Django application to a Go-based microservices architecture requires robust mocked data that matches the live environment. This document outlines a strategy to capture live JSON and RabbitMQ messaging data directly from the legacy application's live machine, and how to utilize this scraped data to automate the creation of accurate mocks for the new Go microservices.

## 2. Approach: Scraping Live Application Data
To securely scrape the existing application, we will use a Python script similar to the automated test scripts found in `original_repo/pum/test`. The script will authenticate with the live backend using HTTP sessions to retrieve the required CSRF tokens and Session IDs, and then iterate through the necessary endpoints.

### 2.1 Accessing the Live System
As the live machine might not be publicly accessible:
1. **VM Snapshot/SSH Access:** Obtain SSH keys stored on the VM snapshot to connect to the working machine.
2. **Script Execution:** We can run the scraping script directly on the live machine via SSH, capturing the JSON output directly to files on the VM, and then using `scp` or `rsync` to pull the files back to the local development environment.
3. Alternatively, if local development machines can establish an SSH tunnel or VPN to the live machine, the script can be run locally pointing to the forwarded port (`localhost:8080` -> `live-arm-ip:80`).

### 2.2 Scraping Script Behavior
The scraping script will utilize `aiohttp` or the standard `requests` module to:
1. Navigate to `http://<live-ip>/accounts/login_old/`.
2. Capture the `csrftoken` from cookies.
3. POST the login credentials (e.g., username/password configured via `.env` files).
4. Save the resulting `sessionid` and `csrftoken`.
5. Iterate through target endpoints and save responses to a local `mocks/` directory.

Target endpoints to scrape include (but are not limited to):
- `/data/ipmi/list?json=true` (IPMI node list)
- `/data/switch/?json=true` (Switch list)
- `/accounts/currentuser?json=true` (Current user details)
- `/core/settings?json=true&json_object_list` (Global settings)
- Any other endpoints that the frontend expects.

## 3. Mock Data Structure & Generation
Once the JSON data is pulled, it will be injected into our Go mock environment.
Currently, our Go app utilizes a 3-tier mocking strategy for external dependencies configured in `system.yaml`.

### 3.1 Static HTTP Mocks
The saved JSON files (e.g., `ipmi_list.json`, `switch_list.json`) can be placed in a dedicated mock data folder (e.g., `pkg/mockdata/` or served directly by the mock endpoints in the individual services).
- For local Go testing, HTTP mock servers (e.g. inside `apptron/cmd/pum-admin/main.go` or specific microservices) will read these JSON files from disk and return them when the legacy compatibility URLs are requested.
- We can write a quick utility script (`cmd/mockgen`) that reads the scraped JSON and generates the appropriate Go struct definitions or SQL seed data if we intend to populate the local SQLite databases.

### 3.2 Dynamic Messaging/RabbitMQ Analysis
The legacy system relies heavily on RabbitMQ for hardware control and alarm propagation. Based on analysis of the original test scripts (`test_asu_x.py`, `test_asu_x2.py`), the message packets have distinct schemas.

**Switch/Link Control Packets (`asuiks_channels1` / `asu-iks-queue`):**
*Create Request:*
```json
{
  "idRoute": 143,
  "reqtype": "create",
  "sw_name1": "test-client-rack1", "portId1": "19", "portName1": "01",
  "sw_name2": "test-client-rack2", "portId2": "20", "portName2": "01",
  "timeout": 200
}
```
*Delete Request:*
```json
{
  "idRoute": 138,
  "reqtype": "delete",
  "timeout": 200
}
```

**Alarm and Status Reports (`asu.iks` exchange):**
```json
{
  "command": "report",
  "timestamp": "2023-12-18 18:55:59.597285+03:00",
  "data": {
    "code": 200,
    "eventId": 916,
    "incomingMessage": {
      "alarmDetails": "{ ... status ... }",
      "alarmId": "15126665",
      "alarmRaisedTime": "2023-12-13T19:18:12.000+03",
      "alarmState": "RAISED",
      "specificProblem": "Отсутствует связь с АСУ ИКС"
    },
    "requestId": 657,
    "validity": "Invalid response body"
  }
}
```

#### Message Mocking Approach
1. For testing Go microservices that publish to RabbitMQ, the scraped data formats dictate the Go struct models we must create for serialization (e.g., `AsuIksRouteRequest`, `AsuIksReport`).
2. We can develop a Go-based "Mock Hardware Controller" that binds to these queues, consumes the `create`/`delete` requests, and automatically publishes a corresponding `report` response to simulate the full lifecycle of a hardware interaction.

## 4. Knowledge Extraction Reference

During the investigation of the original repository (`pum/test` directory), the following key information was extracted which will be essential when re-implementing logic in Go:

### 4.1. Core Application Variables & Definitions
- **LDAP Groups:** `tsumadm` (Admin), `trafadm` (Traffic Admin), `devadm` (Device Admin), `user` (Regular User).
- **Global Settings Keys (`/core/settings`):** `CONTROL`, `TSUM`, `ARM ADDRESS`, `FOREIGN ARMs`, `SPDU`, `STATE`, `ALARMS`, `MAP_SERVER`, `SWITCH_NETWORK`, `ZABBIX_NODEGROUP_PREFIX`, `MAP_REGION`, `ZABBIX_SEVERITY`, `ZABBIX_MAP`, `ZABBIX_TRANSPORT`, `ZABBIX_CONTROL`, `graylog_url_setting`.

### 4.2. Target API Endpoints and Payloads
These HTTP endpoints return specific nested structures that the frontend relies on. The scraping script should fetch and preserve these arrays/objects:
- `/data/ipmi/list?json=true` returns a JSON object. We expect arrays of data where an IP is found at `response['data'][0]['interface']['ip']`.
- `/data/switch/?json=true` returns a JSON object. We expect arrays where an IP is found at `response['switch_list'][0]['ip']`.
- `/data/ipmi/{ip}/status` triggers an async task and returns a `task_id`.
- `/data/switch/{ip}/status1` triggers an async task and returns a `task_id`.
- `/tasks/viewtasks/` lists all tasks and statuses (e.g. `COMPLETED`, `FAILED`).
- `/accounts/currentuser?json=true` returns the current user profile including `id`.
- `/core/settings?json=true&json_object_list` returns system configurations, primarily parsed by extracting objects from the `object_list` array matching a specific `key`.

### 4.3. RabbitMQ Queues & Exchanges
The application uses specific queues and exchanges to communicate with background processors and hardware. When implementing Go RabbitMQ consumers/producers, these exact names must be mapped:
- **ASU IKS Control:** `asu-iks-queue` (or `asuiks_channels1` as noted in some tests)
- **ASU IKS Exchange:** `asu.iks` (Used for publishing incoming event reports/alarms)
- **Switch Checker:** `netdevcheker`
- **IPMI Config:** `ipmictrl`
- **Switch Config:** `switch-ctrld`
- **IP Config Service:** `ipservice`
- **PDU Control:** `netdevpdu`
- **Service Routes:** `SSttrreel1A`
- **SPDU Queues:** `spdu-1`, `spdu-2`
- **Sentry Readiness:** `sentryd-p-factor`
- **Inter-ARM Exchange:** `amqp.fanout`

### 4.4. Backend Architecture and Data Models Reference
By analyzing the legacy Django application (`original_repo/pum/backend`), the following data flow patterns were discovered, which should guide the implementation of Go handlers:

- **`/accounts/currentuser` (User Profile):**
  Constructed from the `CustomUser` model and the LDAP extension. The returned JSON object always includes a list of `groups` and `permissions`. Example response schema structure:
  `{"id": 1, "username": "admin", "groups": ["tsumadm", "devadm"], "permissions": [...]}`

- **`/tasks/viewtasks/` (Background Task Polling):**
  In the legacy system, tasks triggered by background actions (like `/data/switch/{ip}/status1`) return a task ID. The frontend polls `/tasks/viewtasks/` which internally queries a `RedisStore(TaskRecord)` and serializes matching task objects into JSON (`status`, `started`, `id`, `username`, etc). Our Go application needs to replicate this schema, perhaps using an in-memory or SQLite-backed task store instead of Redis.

- **`/data/ipmi/list` (Hardware/IPMI Source):**
  Instead of hitting a database table, this endpoint queries the `ZabbixAPI` directly (via `host.get` and `hostinterface.get` looking for `ipmi_available`). It aggregates the Zabbix result into an array of objects where each object contains an `interface` nested object containing the IPMI IP. When creating the Go mock for this, the mock data structure must perfectly mimic this Zabbix-aggregated format, not a flat database row.

- **`/data/switch/` (Switch Listing):**
  Returns a serialized array of `Switch` models filtered by `GlobalSettings` (`MAP_REGION`) and `Gw` nodes. The serialization is done via a custom `jsonify()` function across the entire model structure. Mocking this involves providing a deeply nested JSON object representing the Switch state, Freeports, and interconnected nodes.

## 5. Mock Data for External Sources

To decouple the Go application from the live infrastructure, we need to gather realistic data payloads from external sources. The Go application uses a 3-tier mocking strategy (real, emulated, mock). The "emulated" tier requires running local mock servers that replicate the responses of GLPI and Zabbix APIs.

### 5.1 Zabbix API Emulation Data
The original application uses `ZabbixAPI` directly to fetch equipment, IPMI, and network maps. An emulated Zabbix API must support JSON-RPC 2.0 requests to the `/api_jsonrpc.php` endpoint with the following methods:

- `apiinfo.version`: Returns the Zabbix API version (e.g., `5.0.x`).
- `host.get`: Returns an array of host objects. The application checks the `name`, `hostid`, and `ipmi_available` fields.
- `hostinterface.get`: Returns interface configurations for hosts. Important fields to mock are `type` (`3` for IPMI), `available`, `hostid`, `dns`, `port` (e.g., `623`), and `main` (`1`).
- `map.get`: Retrieves topological maps (`zabbix_map` from global settings). The emulation must return map objects with `sysmapid`, `name`, `selements` (switches), `links` (connections), `elements` (host relations), and lines.

### 5.2 GLPI API Emulation Data
The application fetches DataCenter locations from GLPI to render maps. An emulated GLPI REST API must expose the following routes with corresponding JSON data:

- `/search/DataCenter?range=0-0`: Returns pagination data like `{"totalcount": <number>}`.
- `/DataCenter?range=0-N`: Returns an array of DataCenter objects containing `id`, `name`, and `locations_id`.
- `/search/Location?range=0-0`: Returns location pagination data.
- `/Location?range=0-N`: Returns an array of Location objects containing `id`, `completename`, `longitude`, and `latitude`.

### 5.3 RabbitMQ Packet Emulation Data
Message payloads need to be serialized as JSON. See section 3.2 for exact schemas for `create`, `delete`, and `report` packets used by `asu-iks-queue` and `asu.iks`.
Additional queues (e.g., `netdevcheker`, `ipmictrl`, `switch-ctrld`) will require capturing live packets during the scraping phase to build out exact Go struct representations.

### 5.4 LDAP Emulation Data
The original application uses `ldap3` in Python to manage authentication and role-based access. It connects to up to two LDAP servers (`AUTH_LDAP_1_SERVER_URI`, `AUTH_LDAP_2_SERVER_URI`) and searches specific organizational units.
The Go LDAP emulation must construct a directory tree with the following characteristics:
- **Users Search Base:** `ou=Users,dc=strela`
  - Object filter: `(&(objectClass=person)(name=<username>))`
  - Example user: `cn=admin,ou=Users,dc=strela`
- **Groups Search Base:** `ou=Groups,dc=strela`
  - Object filter: `(objectClass=posixGroup)`
  - Group attributes: `cn` (e.g., `tsumadm`, `devadm`, `trafadm`, `user`), and `memberUid` (a list of usernames belonging to the group). The application checks if the user's name is inside `memberUid` to assign permissions.

## 6. Implementation Strategy for Emulated External Sources

In the Go repository, external API interactions are governed by `pkg/external/factory.go` configured via `system.yaml`.

### 6.1 Creating HTTP API Emulators (Zabbix & GLPI)
1. **Mock Data Generation:** During the SSH scraping phase (Step 7), run a proxy or script to dump raw JSON responses from the live Zabbix and GLPI APIs into static JSON files (e.g., `zabbix_hosts.json`, `glpi_datacenters.json`).
2. **Go Emulators:** Within the `apptron/cmd/pum-admin/main.go` local mock environment, set up simple HTTP multiplexers (using standard library `net/http` or `gin`) bound to local ports (e.g., `:8081` for Zabbix, `:8082` for GLPI).
3. **Zabbix RPC Handler:** Create a single endpoint `/api_jsonrpc.php` that decodes the JSON-RPC request body, inspects the `method` string, and returns the corresponding static JSON file loaded from disk.
4. **GLPI REST Handler:** Create basic GET route handlers for `/DataCenter` and `/Location` that return static JSON arrays.
5. **Configuration Update:** Update `system.yaml` to point the `external_modules.zabbix` and `external_modules.glpi` endpoint URLs to `http://localhost:8081` and `http://localhost:8082` respectively, while keeping the `mode` set to `emulated`.

### 6.2 Creating RabbitMQ Emulators (Hardware Control)
1. **Queue Simulation:** Instead of HTTP endpoints, hardware controllers and BPKGW mock instances communicate over AMQP. Use a local Docker instance of RabbitMQ for development (`docker run -d -p 5672:5672 rabbitmq`).
2. **Hardware Controller Daemon:** Create a simple Go binary or goroutine inside the `apptron` local test harness that connects to `amqp://localhost:5672`, declares the necessary queues (e.g., `asu-iks-queue`), and consumes requests.
3. **Response Emulation:** When the mock hardware daemon receives a link `create` request, it waits 1-2 seconds, and publishes a `report` JSON payload to the `asu.iks` exchange with an `"eventId"` and `"alarmState": "CLEARED"`, simulating a successful physical switch configuration.

### 6.3 Creating LDAP Emulators (Identity/Roles)
1. **Mock LDAP Provider:** In the Go `identity` service, utilize the existing mock LDAP provider (`services/identity/ldap/mock.go`).
2. **Directory Configuration:** Instantiate an in-memory LDAP directory that defines a root of `dc=strela`.
3. **Object Seeding:** Seed it with `ou=Users` containing test user entries (e.g. `cn=admin`), and `ou=Groups` containing the `posixGroup` definitions (`tsumadm`, `devadm`). Ensure the `memberUid` attributes correctly map back to the test usernames so the frontend correctly renders role-filtered GUI elements.

## 7. Data Structures: Inferable vs. Live Extraction
To efficiently build the Go mock models, it is crucial to understand what data can be directly implemented in Go right now based on source code analysis, versus what data must wait for the live scraping script output.

### 7.1 Data Structures Fully Inferable from Source
The following schemas and structures can be implemented in Go structs and emulators immediately:
- **RabbitMQ Link Control Packets (`asu-iks-queue`):** The exact schema is known (`idRoute`, `reqtype`, `sw_name1`, `portId1`, `portName1`, etc.).
- **RabbitMQ Report Packets (`asu.iks`):** The top-level schema is known (`command`, `timestamp`, `data.code`, `data.eventId`, `data.incomingMessage`).
- **LDAP Directory Structure:** The exact required OUs (`ou=Users`, `ou=Groups`), object classes (`posixGroup`, `person`), and required group `cn` names (`tsumadm`, etc.) are known.
- **Background Task Records (`/tasks/viewtasks/`):** The schema structure includes `status`, `started`, `id`, and `username`.
- **Zabbix / GLPI API Contracts:** We know exactly which JSON-RPC methods (`host.get`, `map.get`) and REST paths (`/DataCenter`) the frontend calls, and exactly which specific JSON keys it looks for in the response (e.g. `ipmi_available`, `sysmapid`, `completename`).

### 7.2 Data Structures Requiring Live Extraction
The following structures are either too dynamic or deeply nested in the legacy Python application to reliably map from source alone. The Go structs for these should only be finalized *after* reviewing the scraped JSON output:
- **Switch Listing Hierarchy (`/data/switch/`):** The legacy app uses a custom `jsonify()` function that serializes the `Switch` Django model along with its nested `Freeports` and interconnected `Gw` nodes. The exact shape of this tree must be observed from a live `/data/switch/?json=true` response.
- **Topological Data:** The actual real-world values for Zabbix `map.get` topologies (e.g., how the `links` array maps to `selements`) and GLPI geographic coordinate data.
- **RabbitMQ Opaque Payloads:** The nested JSON string inside `alarmDetails` on the `asu.iks` queue, and the payloads for untested queues (like `netdevcheker`, `ipmictrl`, `switch-ctrld`).

## 8. Summary of Execution Steps
1. Provision SSH access to the live VM.
2. Develop `scrape_legacy.py` based on `common.py` from the legacy test suite.
3. Execute the script on the live VM and export the resulting `.json` files.
4. Update Go microservices to deserialize this data to populate local SQLite databases, or configure our mocked backend endpoints to serve this static JSON directly.
5. Implement RabbitMQ struct models in Go based on the analyzed packet structures to handle message-based mocks.