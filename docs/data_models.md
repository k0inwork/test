# Data Models Analysis

This document explains the core data structures and Django models used in the system.

## 1. Product (PDU) Management
The `products` app manages the primary entities the system controls.

### 1.1. Product Model (`products.models.Product`)
Represents a physical or virtual device (PDU, OUs).
- **Core Fields:**
  - `name`, `pouType`, `sequential_number`, `region`: Identify the device.
  - `state`: Current operational status.
  - `geo`, `long`, `lat`: Geographical location for mapping.
  - `glpi_uuid`: Linkage to the GLPI CMDB.
- **Sync Logic:** Includes methods for bidirectional sync with GLPI, handling complex naming conventions (e.g., `REGION/#seq-POU name`).

### 1.2. Requests Model (`products.models.Requests`)
Tracks asynchronous requests sent to devices or other modules.
- **Fields:** `request`, `response`, `pending`, `error`, `request_timeout`.
- **Purpose:** Provides an audit trail and state tracking for operations that may take time to complete.

## 2. Gateway and Session Management
The `gws` app handles network tunnel orchestration.

### 2.1. Gw Model (`gws.models.Gw`)
Represents a network gateway.
- **Fields:** `address`, `region`, `state`.
- **Logic:** Connects to physical or virtual gateways (often running the `bpkgw` service).

### 2.2. Session Model (`gws.models.Session`)
Represents a VXLAN tunnel between two gateways.
- **Fields:**
  - `gw1`, `gw2`: The two endpoints of the tunnel.
  - `subnet`: The network address assigned to the tunnel.
  - `tx_bytes`: Traffic statistics collected from the gateway.
- **Usage:** Sessions are created dynamically via the `bpkgw` REST API to establish secure tunnels between regions.

## 3. Global Configuration (`core.models.GlobalSettings`)
A key-value store in the database for system-wide configuration.
- **Usage:** Stores settings like the current Master ARM, LDAP configuration, Zabbix prefixes, etc.
- **History:** Uses `django-simple-history` to track changes to critical settings.

## 4. Task Tracking (`tasks.models.TaskRecord`)
While Celery handles the execution, `TaskRecord` provides a persistent record in the database of what tasks were run, by whom, and their results. This is essential for auditing and providing feedback to the UI.
