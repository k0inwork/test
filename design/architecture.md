# System Architecture Design: PUM Rewrite (Go)

## 1. Introduction
This document outlines the architectural transition of the Product Unit Management (PUM) system from a monolithic Django/Celery application to a microservices-based architecture implemented in Go.

## 2. Core Goals
- **Scalability**: Move away from monolithic constraints.
- **Maintainability**: Clear boundaries between domains (Identity, Product, Network, etc.).
- **Performance**: Leverage Go's concurrency model for hardware interaction and background tasks.
- **Simplicity**: Use standard stable libraries and SQLite for initial persistence.
- **Admin Empowerment**: Provide a high-performance, local-first management environment using Apptron.

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
│   ├── terminal/       # Mocked/Stubs
│   └── pum-cli/        # Go-WASM Tool (CLI/TUI)
├── pkg/                # Shared packages
│   ├── common/         # Utils, Middleware, Logging
│   └── models/         # Shared domain models
├── api/                # API definitions (OpenAPI/GraphQL)
├── design/             # Design documentation
└── apptron-docs/       # Apptron integration strategy
```

## 4. Microservices Breakdown

*(Sections 4.1 to 4.6 remain unchanged)*

### 4.7 Apptron Admin Center (New)
The Admin Center is a browser-based, local-first client environment powered by Apptron. It consists of:
- **Headless WASM Core**: A background service in the browser that handles local data state and caching.
- **pum-cli**: A WASM-compiled Go tool providing both a CLI for scripting and a TUI for interactive management.
- **Web Dashboards**: HTML/JS views integrated into the Apptron workspace.
- **Bridge Agent**: A local native Go agent that bridges the browser's virtual network to physical hardware.

## 5. Technology Stack
- **Language**: Go (Golang) 1.25+
- **Web Framework**: Gin Gonic (for REST)
- **WASM Client Framework**: Bubbletea (for TUI), syscall/js (for Headless Core)
- **Database**: SQLite (local file per service)
- **Admin Environment**: Apptron (x86 emulation + WASM)

## 6. Communication Patterns
- **Inter-service**: RESTful HTTP calls.
- **Admin to Server**: REST/GraphQL over HTTPS.
- **Admin to Hardware**: Tunneled TCP via Bridge Agent and Apptron Virtual Network.

## 7. Mocking Strategy
*(Remains unchanged)*
