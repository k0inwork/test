# Architecture Evolution: From Django to Go

## Legacy Architecture (PUM-Django)
The original PUM implementation was built as a monolithic (or semi-distributed) Django application.

-   **Backend**: Python/Django 4.0.
-   **Communication**: RabbitMQ with a custom fanout/RPC layer for module interaction.
-   **State Management**: Centralized PostgreSQL database with heavy use of Redis for volatile state and distributed locking.
-   **Frontend**: Traditional Server-Side Rendering (SSR) with templates and integrated JavaScript (e.g., xterm.js for WebSSH).
-   **Strengths**: Fast initial development, rich ecosystem of Python libraries.
-   **Challenges**: Performance bottlenecks in high-concurrency scenarios, complex deployment of many moving parts (Celery, RabbitMQ, multiple workers).

## Modern Architecture (PUM-Go)
The new architecture shifts toward a more performant, scalable, and manageable ecosystem.

-   **Backend**: Specialized microservices written in Go.
-   **Communication**: Lightweight inter-service communication (currently transitioning to a unified registry-based model).
-   **State Management**: Microservice-specific persistence, with a focus on local-first capabilities in the client.
-   **Frontend**: The Apptron Rich Client. Instead of simple pages, the frontend is an "OS in the browser" that hosts WebViews and WASM applications.

### Key Architectural Shifts

| Feature | Legacy (Django) | Modern (Go/Apptron) |
| :--- | :--- | :--- |
| **Language** | Python | Go |
| **UI Model** | Server-Side Templates | Apptron WebViews + WASM/WASI |
| **Tooling** | Browser-based JS Tools | Go-WASM binaries (CLI/TUI) |
| **Terminal** | Django-WebSSH (Proxy) | WASM-based (Local Execution) |
| **Latency** | Network Dependent | Local-First (WASM) |
| **Deployment** | Complex (Django+Celery+RMQ) | Simplified (Go Binaries + Assets) |

## The Role of Apptron
Apptron serves as the bridge between these worlds. It provides the runtime environment that allows us to ship native-like management tools (WASM) alongside traditional web interfaces (Dashboards), all within a single unified application shell.
