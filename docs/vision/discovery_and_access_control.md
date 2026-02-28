# Dynamic Discovery and Capability-Based Access Control

## Overview
The PUM-Go architecture is built for extensibility. Instead of a hard-coded set of features, the system relies on **Dynamic Discovery** and **Capability-Based Access Control** to determine what tools and data are available to any given user.

## Dynamic Service Discovery
The **Registry Service** acts as the source of truth for the entire ecosystem.
-   **Registration**: When a microservice (e.g., Inventory, Network) starts, it registers its endpoint and its **Capabilities** with the Registry.
-   **Capabilities**: These are machine-readable tags (e.g., `users`, `nodes`, `port-control`, `network-manage`) that define what functions the service provides.
-   **Heartbeats**: Services maintain an active status via regular heartbeats, ensuring that the UI only presents available features.

## Capability-Based UI (The "Ghost" Interface)
Both the Go-based frontend and the Apptron rich client use the Registry to build their interfaces dynamically.
-   **Feature Detection**: If the "network-manage" capability is not present in the Registry, the corresponding menus, dashboards, and CLI commands are hidden.
-   **Resilience**: The UI gracefully degrades or upgrades based on the real-time availability of backend services.

## The Future of Access Control (RBAC)
We are evolving toward a decentralized security model that links User Roles to Service Capabilities.

### 1. User Roles
Users are assigned roles (e.g., `Admin`, `Operator`, `ReadOnly`).

### 2. Capability Mapping
An Access Control service (integrated with Identity) will manage the mapping between roles and capabilities.
-   `Admin` -> All Capabilities.
-   `Operator` -> `nodes`, `port-control`, `inventory`.
-   `ReadOnly` -> `nodes-view`, `inventory-view`.

### 3. Enforcement
-   **Frontend Enforcement**: The UI queries the Access Control service to determine which discovered capabilities the *current user* is authorized to use. Elements that the user cannot access are removed from the DOM or WASM runtime.
-   **Backend Enforcement**: The "Bridge" in Apptron and the individual microservices verify the user's session and role against the requested capability before executing any action.

## Strategic Goal
The goal is a "Pluggable Command Center" where adding a new management capability is as simple as deploying a new microservice. The system will automatically discover the service, map it to the appropriate user roles, and update the GUI/CLI/TUI interfaces across all platforms.
