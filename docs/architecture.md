# System Architecture Overview

This document provides a high-level overview of the system's architecture, components, and their interactions.

## 1. Core Components

The system is a distributed management platform designed for controlling network devices, gateways, and virtual machines. It consists of the following primary components:

### 1.1. PUM (Product Unit Management) Backend
- **Framework:** Django 4.0.1
- **Role:** Central management hub, provides the web interface and REST/GraphQL APIs.
- **Responsibilities:**
  - User authentication and authorization (via LDAP/Kerberos).
  - High-level business logic for product (PDU) and gateway management.
  - State management for Multi-Master configurations.
  - Coordination of background tasks.

### 1.2. Message Broker (RabbitMQ)
- **Role:** Communication backbone for inter-module and inter-ARM coordination.
- **Key Exchanges:**
  - `amqp.fanout`: Used for broadcasting state changes and commands across multiple ARM instances.
  - Module-specific queues: Used for communicating with specialized controllers (e.g., `netdevcheker`, `switch-ctrld`).

### 1.3. Task Queue & Result Backend (Redis)
- **Broker for Celery:** Handles asynchronous task distribution.
- **Django Channels Layer:** Manages WebSocket connections for real-time UI updates.
- **Shared State Storage:** Custom abstractions (`RedisStore`, `RedisVar`) store volatile state shared across different Python processes.

### 1.4. Network Gateways (bpkgw)
- **Framework:** Flask
- **Role:** Low-level network interface management.
- **Responsibilities:**
  - Creating and destroying VXLAN tunnels using `ip link` commands.
  - Monitoring traffic (TX bytes) on tunnel interfaces.
  - Providing a REST API for session management.

## 2. Communication Flow

### 2.1. Internal (Django to Workers)
The Django application dispatches tasks to Celery workers via Redis. These workers handle long-running operations like synchronization with external systems or complex network configurations.

### 2.2. Inter-Module (RabbitMQ)
For communication with external hardware controllers or other ARM instances, the system uses RabbitMQ. A custom `PikaChannelSingleton` ensures efficient connection management, while the `event_consumer` library allows Celery workers to react to incoming RabbitMQ messages.

### 2.3. Multi-Master Coordination
The system supports multiple "Master" instances. They coordinate their state (who is currently the active Master) by broadcasting messages over the `amqp.fanout` exchange in RabbitMQ. Redis is used to track which requests have been processed.

## 3. External Integrations
- **LDAP/Kerberos:** Centralized authentication.
- **GLPI:** Used as a CMDB for syncing product/datacenter information.
- **Zabbix:** Monitoring of devices and network status.
- **QEMU:** Virtualization management, typically used for gateway or testing environments.
- **Graylog:** Centralized logging for system events.
