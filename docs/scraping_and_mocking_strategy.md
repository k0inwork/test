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

## 4. Summary of Execution Steps
1. Provision SSH access to the live VM.
2. Develop `scrape_legacy.py` based on `common.py` from the legacy test suite.
3. Execute the script on the live VM and export the resulting `.json` files.
4. Update Go microservices to deserialize this data to populate local SQLite databases, or configure our mocked backend endpoints to serve this static JSON directly.
5. Implement RabbitMQ struct models in Go based on the analyzed packet structures to handle message-based mocks.