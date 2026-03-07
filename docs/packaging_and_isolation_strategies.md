# Microservices Packaging and Isolation Strategies

This document analyzes different strategies for packaging and isolating the Go microservices architecture (including external dependencies like Jaeger and Prometheus) across macOS and Linux, without relying on Docker or heavy virtualization.

The goal is to move away from a simple `run_all.sh` script to a more robust solution that provides better process management and, ideally, security/filesystem isolation.

---

## 1. Docker / OCI Containers (Baseline Examinable Option)

While the initial constraint was to avoid Docker or heavy virtualization, Docker remains the industry standard baseline against which all other options are measured.

### How it Works
*   **Packaging:** We write a `Dockerfile` (or `docker-compose.yml`) that defines the entire runtime environment, pulling in Alpine Linux or Distroless images, copying compiled Go binaries, and orchestrating external images (like Jaeger and Prometheus).
*   **Isolation (Process & Network):** Docker provides native network namespaces. Microservices can communicate on an internal virtual network without exposing ports to the host OS.
*   **Environment Security:** On Linux, Docker relies on `cgroups` and standard Linux namespaces. It is highly secure, provided the containers do not run as root.

### Effort: Low
*   Docker compose is trivial to write and maintain. Almost all external dependencies (Jaeger, Prometheus) provide official, optimized Docker images.

### Consequences
*   **Pros:**
    *   **Universal Packaging:** Solves the packaging problem completely. A single `docker-compose up` works everywhere.
    *   **External Binaries:** Seamlessly handles Jaeger and Prometheus without needing custom native orchestrators.
*   **Cons (The Dealbreakers):**
    *   **macOS Overhead:** Docker on macOS *must* run a hidden Linux VM (via HyperKit, Lima, or qemu) to utilize `cgroups`/namespaces. This introduces significant CPU/Memory overhead and battery drain, completely violating the lightweight, native execution preference.
    *   **Filesystem Sync:** Mounting local macOS directories into the Linux VM for live-reload development is notoriously slow.

---

## 2. WebAssembly (WASM / WASI)

Given the project's existing use of WebAssembly (Apptron/Wanix), compiling the Go microservices to WASM and running them in a native WASM runtime (like Wazero, Wasmtime, or Wasmer) is a strong candidate for true cross-platform isolation.

### How it Works
*   **Packaging:** Go microservices are compiled to `.wasm` binaries (`GOOS=wasip1 GOARCH=wasm go build`). These are entirely platform-independent.
*   **Isolating:** WASI (WebAssembly System Interface) operates on a capability-based security model. By default, a WASM module has **zero** access to the host file system, network, or environment variables. The host runtime must explicitly grant access to specific directories or network sockets.

### Effort: High
*   **Go Compilation:** Go 1.21+ has excellent support for `wasip1`, but any CGO dependencies (like `go-sqlite3`) will fail to compile. The project currently uses SQLite. We would need to migrate to a pure Go SQLite implementation (like `modernc.org/sqlite`).
*   **Networking Constraints:** Standard Go `net/http` servers rely on full networking stacks. WASI's networking support is still evolving (WASI Preview 2/3). While basic HTTP servers can work in runtimes like Wasmer/Wasmtime, complex networking, raw sockets, or background goroutines interacting with external services might require significant workarounds or custom host functions.
*   **External Binaries:** We cannot compile pre-existing binaries like Jaeger or Prometheus to WASM easily. They would still need to run natively on the host OS. The WASM orchestrator would need to manage both WASM modules (for our Go code) and native processes (for Jaeger/Prometheus).

### Using an External Native Module (Host Orchestrator/Capability Provider)
A very common and reasonable pattern for complex WASM architectures is the **Host Orchestrator Pattern** (using Host Functions).
Instead of forcing the WASM runtime to handle raw TCP/IP or CGO SQLite calls, we build a single, native "Host Runner" (written in Go) that runs on the OS.

*   **Networking (HTTP/TCP):** The native Host Runner binds to the actual OS ports. When an HTTP request comes in, the Host Runner passes the request payload into the WASM module. The WASM module processes the logic and returns a response payload, which the Host Runner sends back over the real network.
*   **Database (SQLite):** The Host Runner holds the actual CGO SQLite connection. The WASM module makes RPC-like calls (Host Functions) to the native runner (e.g., `execute_query("SELECT * FROM users")`), and the native runner executes the CGO code and passes the result back into WASM.

**Is this reasonable?**
Yes, this is highly reasonable and is the standard industry approach for enterprise WASM (used by platforms like Fermyon Spin or Cloudflare Workers).
*   **Pros:** It completely solves the CGO/SQLite limitation and bypassing immature WASI networking. The WASM modules remain pure, portable business logic.
*   **Cons:** It increases architecture complexity. We have to design an RPC interface or use something like `wazero` to define strict Host Functions that bridge the gap between the native OS Go Runner and the internal WASM Go logic.

### Advanced IPC: Streaming over Unix Sockets / Pre-opened FDs via WASI
Instead of writing custom, proprietary Host Functions for every single operation (e.g., `execute_sqlite_query()`) or implementing a complex virtual filesystem protocol like 9P, a much simpler and highly performant alternative is to stream raw bytes over **Unix Sockets** (or pipes) passed directly into the WASM environment via WASI.

*   **How it Works:** Before launching the WASM module, the native Go Host Runner opens standard OS-level Unix sockets (or anonymous pipes). It then configures the WASM runtime (like `wazero`) to "pre-open" these file descriptors (FDs) and map them directly into the WASM module's virtual filesystem (e.g., mapping FD 3 to `/mnt/host_rpc.sock`).
*   **What it Solves:**
    *   **High-Performance IPC:** The WASM module simply opens `/mnt/host_rpc.sock` and performs standard Go `io.Reader` and `io.Writer` stream operations. No network stack is required inside WASM, and no complex filesystem protocols (like 9P) are needed. It's just raw byte streaming between the guest and the host.
    *   **Universal RPC over Sockets:** The Go Host Runner listens on the other end of the socket. The WASM module writes JSON, Protobuf, or standard RPC payloads to the socket, asking the Host Runner to execute complex tasks (like SQLite queries or HTTP outbound requests). The Host Runner executes them natively and streams the results back through the socket.
    *   **Portability:** The WASM module remains 100% standard `GOOS=wasip1`. It only needs standard file I/O (`os.Open`, `os.Read`, `os.Write`). It doesn't need to know about custom host functions or network stacks.
*   **Pros:** Extremely low overhead, natively supported by the standard WASI spec (pre-opened FDs), and significantly less architectural boilerplate than a full 9P virtual filesystem. The WASM modules become completely decoupled from the specific host runtime environment.
*   **Cons:** We have to define our own lightweight framing/RPC protocol over the raw stream (e.g., prefixing messages with a length header so the receiver knows when a full request/response has arrived).

**Addressing Telemetry across the Socket:**
A major concern with defining a custom RPC protocol over Unix sockets is that internal requests (like an HTTP call made by the WASM module) might become "hidden" from the global distributed tracing (Jaeger).
*   **The Solution:** OpenTelemetry (OTel) context propagation solves this completely.
*   The project already uses standard OTel instrumentation. When the WASM module initiates a request over the socket, it must serialize the current OTel span context (specifically the W3C `traceparent` and `tracestate` headers) into the metadata of the custom RPC payload.
*   When the native Go Host Runner receives the RPC payload from the socket, it extracts these W3C headers, creates a new child span natively, and executes the complex task (like SQLite). This ensures a single, unbroken trace in Jaeger from the initial HTTP entry point, through the WASM boundary, across the Unix socket, and into the native database/network execution.

### Consequences
*   **Pros:**
    *   **Ultimate Security Sandbox:** True, strict isolation on both macOS and Linux without any OS-level containers.
    *   **Portability:** A single `.wasm` file runs exactly the same anywhere.
    *   **Synergy:** Aligns with the project's existing Apptron/WASM architecture.
*   **Cons:**
    *   Significant refactoring required for CGO dependencies (SQLite).
    *   Networking in WASI is immature; complex HTTP microservices might encounter subtle bugs or missing features.
    *   Performance overhead compared to native Go binaries.
    *   Does not solve the packaging/isolation problem for Jaeger and Prometheus.

---

## 3. Platform-Specific Sandboxing Wrapper

This approach creates a smart Go-based launcher that uses the native security features of the underlying OS to sandbox the native Go binaries.

### How it Works
*   **Packaging:** Distribute native, pre-compiled Go binaries for `darwin/amd64`, `darwin/arm64`, and `linux/amd64`.
*   **Isolating:**
    *   **Linux:** The launcher wraps the service execution in `bwrap` (Bubblewrap) or `firejail`, mounting only necessary directories read-only and exposing specific ports.
    *   **macOS:** The launcher uses the native `sandbox-exec` utility. We write `.sb` (Scheme) profile files that explicitly deny network and filesystem access, except for whitelisted paths and ports.

### Effort: Medium
*   **Development:** Requires writing and maintaining separate sandbox profiles for Linux (`bwrap` arguments) and macOS (`.sb` files).
*   **macOS Complexity:** Apple considers `sandbox-exec` deprecated and its `.sb` syntax is undocumented and subject to change, though it is still widely used under the hood by macOS apps.

### Consequences
*   **Pros:**
    *   Achieves strong filesystem and network isolation on both platforms without Docker.
    *   Zero performance penalty; services run as native binaries.
    *   Can easily wrap external binaries like Jaeger and Prometheus in the same sandbox profiles.
*   **Cons:**
    *   High maintenance burden: We have to maintain two completely different sandboxing paradigms (macOS vs Linux).
    *   macOS `sandbox-exec` is fragile and could break in future OS updates.

---

## 4. Lightweight Process Manager (Process Isolation Only)

If strict security sandboxing (preventing a compromised service from reading `~/.ssh`) is less important than operational isolation (clean logs, easy restarts, preventing zombie processes), a Process Manager is the simplest upgrade from `run_all.sh`.

### How it Works
*   **Packaging:** Distribute native Go binaries and a `Procfile`.
*   **Isolating:** Use tools like `Goreman` (a Go implementation of Foreman), `Hivemind`, or `Overmind` (which uses tmux under the hood). The manager spawns each service as a child process.

### Effort: Very Low
*   Create a `Procfile` mapping service names to their start commands (e.g., `identity: go run services/identity/main.go`).
*   Run `goreman start`.

### Consequences
*   **Pros:**
    *   Immediate improvement over `run_all.sh`.
    *   Beautiful, multiplexed, color-coded logging.
    *   Guaranteed clean shutdown (sends SIGTERM to the whole process group, eliminating zombie processes).
    *   Easily manages Jaeger and Prometheus alongside Go services.
*   **Cons:**
    *   **Zero Security Isolation:** Services run as the current user with full access to the filesystem and network.
    *   Does not solve the "packaging" problem (dependencies still need to be installed on the host).

---

## 5. Nix (Nixpkgs / Flakes)

Nix provides absolute environment isolation and reproducible builds without containerization.

### How it Works
*   **Packaging:** We write a `flake.nix` that defines the exact compiler versions, dependencies, and binaries needed (Go, Jaeger, Prometheus).
*   **Isolating:** Nix runs everything in a specific, isolated environment closure. The application cannot see dependencies outside its explicitly defined graph.

### Effort: Medium to High
*   Requires learning the Nix language to write derivations for our Go microservices and fetching Jaeger/Prometheus.

### Consequences
*   **Pros:**
    *   **Perfect Reproducibility:** "It works on my machine" guarantees it works perfectly on every developer's Mac and Linux machine.
    *   Handles packaging elegantly; users just run `nix run .#start-all`.
    *   No Docker needed.
*   **Cons:**
    *   Steep learning curve for developers unfamiliar with Nix.
    *   Provides *environment* isolation (dependencies), not *security* isolation (a running service can still access the host filesystem).

---

## 6. The Hybrid Approach (Different Dev vs. Prod Environments)

If we relax the constraint that the local development environment must have the exact same isolation mechanisms as production, we unlock the most pragmatic, industry-standard approach. Assuming macOS is primarily for development and Linux is strictly for production:

### Development (macOS / Local Linux): Optimize for Speed and Low Overhead
*   **Strategy:** **Lightweight Process Manager** (`Goreman` or `Hivemind`).
*   **How it Works:** Developers run `goreman start`. All Go microservices compile natively and rapidly on the host OS. Jaeger and Prometheus run natively.
*   **Why it's Best Here:** It completely avoids the macOS Docker VM penalty (high CPU/Memory drain, slow filesystem syncs). Developers get instant compilation feedback, beautiful multiplexed logging, and clean process shutdowns without any complex virtualization or sandboxing layers getting in the way.

### Production (Linux Server): Optimize for Strict Isolation and Standard Deployment
*   **Strategy:** **Docker / OCI Containers** (or strictly configured `systemd` + `bwrap`).
*   **How it Works:** CI/CD builds standard Docker images of the natively compiled Go binaries. Production deploys them via `docker-compose` or Kubernetes.
*   **Why it's Best Here:** Since the production server is Linux, Docker runs natively using kernel `cgroups` and namespaces. There is zero VM overhead. It provides perfect, battle-tested network and filesystem isolation. It trivially handles external dependencies (Jaeger/Prometheus) using official images.
*   *Alternative:* If Docker is truly banned even in production, use standard native `systemd` service files wrapped in `bwrap` (Bubblewrap) or `firejail` to achieve similar strong OS-level sandboxing on Linux without a daemon.

---

## 7. Packaging and Execution Tooling (The Selected Path)

Based on the hybrid approach, the tooling to manage these distinct environments should be split between Configuration Management (Ansible) and a Simple Execution Script.

### Packaging & Deployment: Ansible
*   **Production (Linux):** We will use **Ansible** playbooks to handle the actual packaging and deployment to the Linux servers. Ansible is perfectly suited to ensure `bwrap` is installed, write the complex `systemd` unit files, copy the natively compiled binaries, and set up the isolated read-only filesystems.
*   **Development (macOS):** No complex packaging tool is needed. The raw git repository structure combined with dynamic loading (`go run`) is sufficient, as building or pulling the repository is a one-time setup step for a developer.

### Day-to-Day Execution: Simple Wrapper Script
We will replace `run_all.sh` with a simple, intelligent wrapper script (e.g., `start.sh`) focused purely on execution.
*   The script uses `uname -s` to detect the OS.
*   **If `Darwin` (macOS):** It simply executes the Lightweight Process Manager (`goreman start`) against the local repository files.
*   **If `Linux` (Production):** It assumes Ansible has already done the heavy lifting of packaging. The script merely validates the environment and ensures the `systemd` services (wrapped in `bwrap`) are running. This gives developers and operators a unified, simple command to start the application everywhere.

---

## Final Recommendation Summary

1.  **Architecture:** Adopt the **Hybrid Approach**. Use a **Lightweight Process Manager** (like `Goreman`) for local macOS development. For production Linux servers, use native **`systemd` + `bwrap`** for strict security isolation without Docker.
2.  **Tooling:** Use **Ansible** exclusively to package and configure the complex Linux production environment. Rely on the raw git repository structure for macOS development.
3.  **Execution:** Replace `run_all.sh` with a simple **`start.sh`** script that detects the OS and routes to the appropriate local runner or system service.
4.  **Long-Term Goal:** Investigate **WASM/WASI** (using the Host Orchestrator Pattern and Unix Socket IPC) for true, cross-platform security sandboxing without VMs or OS-specific wrappers.