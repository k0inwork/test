# Architecture Evolution: From Django to Go

## Legacy Architecture (PUM-Django)
The original PUM implementation was built as a monolithic (or semi-distributed) Django application.

-   **Backend**: Python/Django 4.0.
-   **Communication**: RabbitMQ with a custom fanout/RPC layer for module interaction.
-   **State Management**: Centralized PostgreSQL database with heavy use of Redis for volatile state and distributed locking.
-   **Frontend**: Traditional Server-Side Rendering (SSR) with Django templates.
-   **Access Control**: Centralized Django RBAC, often tightly coupled with specific views and models.

## Modern Architecture (PUM-Go)
The new architecture shifts toward a more performant, scalable, and manageable ecosystem with a focus on Go and client-side execution.

-   **Backend**: Specialized microservices written in Go.
-   **Communication**: Lightweight inter-service communication via a central **Registry** that tracks service **Capabilities**.
-   **State Management**: Microservice-specific persistence, with a focus on local-first capabilities in the client.
-   **Frontend Service**: A dedicated Go-based web service (`services/frontend`) that serves high-performance Gin templates for standard web access.
-   **Rich Client**: The Apptron environment, hosting WebViews and WASM applications for power users.
-   **Access Control**: Decentralized, capability-based RBAC. Access to GUI elements and API endpoints is dynamically determined by the intersection of user roles and discovered service capabilities.

### Key Architectural Shifts

| Feature | Legacy (Django) | Modern (Go/Apptron) |
| :--- | :--- | :--- |
| **Language** | Python | Go |
| **Discovery** | Hard-coded / Settings | Dynamic (Registry) |
| **Extensibility** | Manual View Updates | Capability-based (Auto-discovery) |
| **Access Control** | Centralized SSR RBAC | Decentralized Capability RBAC |
| **UI Model** | Django SSR | Go SSR (Gin) + Apptron WASM |
| **Terminal** | Django-WebSSH (Proxy) | WASM-based (Client-Side) |

## The Coexistence of SSR and Apptron
The modern architecture doesn't discard server-side rendering. Instead, it uses a Go-based frontend for standard, ubiquitous access while leveraging Apptron for the rich, high-performance "Command Center" experience. This hybrid approach, governed by dynamic discovery, ensures that the system is both accessible and powerful.
