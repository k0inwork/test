# Application Component Summary

This document provides a concise overview of the various apps and modules within the PUM system, highlighting their core responsibilities and technical roles.

## 1. Core Management Platform (Django Apps)

### 1.1. `products`
Manages the lifecycle of **Product Distribution Units (PDUs)** and operational units. Handles integration with GLPI for inventory sync and geographical mapping (GIS) of physical assets.

### 1.2. `gws` (Gateways)
Orchestrates the high-level management of **Network Gateways** and the creation of virtual sessions (VXLAN tunnels) between them. It links users to specific network topologies.

### 1.3. `data` (Network Equipment)
Focuses on the low-level inventory of **Switches and Ports**. It tracks physical connections (`SwitchConnection`) and client assignments (`SwitchCustomerConnection`), and triggers automated configuration updates via specialized modules.

### 1.4. `services` (Connectivity Services)
Provides high-level abstractions for data and key transmission services. It manages "orders" for connectivity (`KeyService`, `DataService`), coordinating the necessary switch and gateway configurations to satisfy client requirements.

### 1.5. `tasks` (Operations Tracking)
A central hub for tracking asynchronous operations. It maintains a persistent record of system tasks (`TaskRecord`), their execution status, and results, providing the data for UI progress bars and audit logs.

### 1.6. `accounts` (Identity & Security)
Handles user management, custom roles, and fine-grained permissions. It integrates with LDAP/Kerberos and implements the `ActivityLog` system for security auditing.

### 1.7. `core` (System Foundation)
Contains the shared library of utilities, base models (`GlobalSettings`), distributed locking mechanisms, and the custom RabbitMQ/Redis communication layers used by all other apps.

### 1.8. `network`
Manages IP address allocation and interaction with the DNS/DHCP controllers. It handles the logic for provisioning logical networks over physical equipment.

## 2. Specialized Controllers & Modules

### 2.1. `bpkgw` (VXLAN Controller)
A Flask-based service that executes low-level Linux networking commands to create, destroy, and monitor **VXLAN tunnels**. It provides the actual "data plane" control.

### 2.2. `control` (State Manager)
A lightweight controller that manages the operational state (START/STOP/REBOOT) and log collection for managed entities via HTTP or RabbitMQ RPC.

### 2.3. `modules/` (Hardware Proxies)
A set of Python modules (IPMI, PDU, Switch, etc.) that act as asynchronous proxies for specialized hardware. They translate high-level Python calls into specific protocol messages (RabbitMQ) for dedicated controllers.

### 2.4. `django_webssh`
Provides an integrated terminal for secure remote access to managed equipment, including session recording for audit compliance.

## 3. Infrastructure Services

- **`ftpserver`**: A dedicated FTP service for managed devices to upload/download configuration backups and logs.
- **`rabbitmq` & `redis`**: The communication backbone for message passing, task queuing, and distributed state caching.
- **`ldap` & `kerberos`**: External identity providers for centralized authentication.
- **`bind` & `isc`**: Underlying services for DNS and DHCP management, controlled via the `ip-module`.
