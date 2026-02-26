# System Architecture Design: PUM Rewrite (Go)

## 1. Introduction
This document outlines the architectural transition of the Product Unit Management (PUM) system from a monolithic Django/Celery application to a microservices-based architecture implemented in Go.

## 2. Core Goals
- **Scalability**: Move away from monolithic constraints.
- **Maintainability**: Clear boundaries between domains (Identity, Product, Network, etc.).
- **Performance**: Leverage Go's concurrency model for hardware interaction and background tasks.
- **Simplicity**: Use standard stable libraries and SQLite for initial persistence.

## 3. Monorepo Structure
The project will follow a monorepo approach for easier management of shared models and inter-service coordination.

```
/
├── cmd/                # Entry points for services
├── services/           # Microservice implementations
│   ├── identity/       # Auth, Users, Roles
│   ├── product/        # Physical Nodes, Equipment
│   ├── network/        # Mocked/Stubs
│   ├── inventory/      # Mocked/Stubs
│   ├── task/           # Mocked/Stubs
│   ├── hardware/       # Mocked/Stubs
│   └── terminal/       # Mocked/Stubs
├── pkg/                # Shared packages
│   ├── common/         # Utils, Middleware, Logging
│   └── models/         # Shared domain models
├── api/                # API definitions (OpenAPI/GraphQL)
└── design/             # Design documentation
```

## 4. Microservices Breakdown

### 4.1 Identity Service
- **Responsibility**: Authentication, Authorization, User management, Roles.
- **Integration**: Initially internal DB, later CAS.
- **APIs**: REST (Login, CRUD Users), GraphQL (Query User profiles).

### 4.2 Product Service
- **Responsibility**: Lifecycle of "Physical Nodes" (previously Products/PDUs), Geo-mapping.
- **Data Model**: Maps to the current `Product` Django model (Node focus).
- **APIs**: REST (CRUD Nodes), GraphQL (Deep queries).

### 4.3 Network & Gateway Service
- **Responsibility**: Routing, IP Allocation, DNS/DHCP.
- **Mocking**: External routing modules and IPAM backends will be mocked.

### 4.4 Inventory Service
- **Responsibility**: Switches, Ports, Physical Connections.
- **Mocking**: External switch configuration modules will be mocked.

### 4.5 Task & Audit Service
- **Responsibility**: System-wide activity logs, Async task tracking.

### 4.6 Hardware Communication Service
- **Responsibility**: Asynchronous proxy for RabbitMQ-accessible modules.
- **Mocking**: RabbitMQ and specific hardware module responses will be mocked.

## 5. Technology Stack
- **Language**: Go (Golang) 1.25+
- **Web Framework**: Gin Gonic (for REST)
- **GraphQL**: gqlgen
- **ORM**: GORM (with SQLite driver)
- **Database**: SQLite (local file per service)
- **Mocking**: Interface-based stubs for un-implemented services.

## 6. Communication Patterns
- **Inter-service**: RESTful HTTP calls (Internal).
- **External API**: REST (JSON) with `?json=true` compatibility (optional) and GraphQL.
- **Frontend**: Similar to existing Bootstrap-based Django templates, potentially moved to a separate SPA or maintained as Go-rendered templates.

## 7. Mocking Strategy
Un-implemented services will be represented by Go interfaces. In Phase 1, the `Product` service will interact with a `MockInventoryService` and `MockNetworkService` that return static or randomized data, allowing for UI and logic development without full implementation.
