# Architecture Evolution: From Django to Go

## Legacy Architecture (PUM-Django)
The original PUM implementation was built as a monolithic (or semi-distributed) Django application.

-   **Backend**: Python/Django 4.0.
-   **Communication**: RabbitMQ using Pika for custom fanout/RPC interaction.
-   **State Management**: Centralized PostgreSQL database with Redis for volatile state and distributed locking.
-   **Frontend**: Traditional Server-Side Rendering (SSR) with Django templates.
-   **Access Control**: Centralized Django RBAC.

## Modern Architecture (PUM-Go)
The new architecture shifts toward a more performant, scalable, and manageable ecosystem with a focus on Go and client-side execution.

-   **Backend**: Specialized microservices written in Go.
-   **Communication Evolution**: Moving from RabbitMQ/Pika to a more performant, lightweight, and Go-native communication bus (e.g., NATS or a registry-based internal bus). This reduces latency and simplifies the infrastructure stack.
-   **Discovery**: Central **Registry** that tracks service **Capabilities** in real-time.
-   **State Management**: Microservice-specific persistence, emphasizing local-first capabilities in the client.
-   **Frontend Service**: A dedicated Go-based web service (`services/frontend`) that serves high-performance Gin templates.
-   **Rich Client**: The Apptron environment, hosting WebViews and WASM applications for power users.
-   **Access Control**: Decentralized, capability-based RBAC.

### Key Architectural Shifts

| Feature | Legacy (Django) | Modern (Go/Apptron) |
| :--- | :--- | :--- |
| **Language** | Python | Go |
| **Communication**| RabbitMQ / Pika | NATS / Internal Bus |
| **Discovery** | Hard-coded / Settings | Dynamic (Registry) |
| **Extensibility** | Manual View Updates | Capability-based (Auto-discovery) |
| **UI Model** | Django SSR | Go SSR (Gin) + Apptron WASM |
| **Terminal** | Django-WebSSH (Proxy) | WASM-based (Client-Side) |

## The Coexistence and Convergence
The transition is governed by the "Strangler Fig" pattern. We maintain the Go-based frontend for standard access while leveraging Apptron for the rich "Command Center" experience. As legacy modules are ported to Go microservices, the Registry dynamically updates the capabilities available to both frontends, ensuring a smooth and incremental convergence.
