# Testing the PUM Admin Distro

## Phase 1: Mock Connectivity Test

The goal of this test is to verify that the WASM client (running in the browser) can communicate with the Bridge Agent (running on the host).

### 1. Build and Run the Admin Center
```bash
bash apptron/scripts/run_phase1.sh
```
This will build the assets and start the `pum-admin` runner on port 8080.

### 2. Verify Health
Open your browser or use curl:
```bash
curl http://localhost:8080/health
```
Should return `OK`.

### 3. Run WASM-to-Bridge Ping Test
1. Open `http://localhost:8080` in your browser.
2. Open the integrated terminal (which runs your custom Alpine bundle).
3. Type `pum ping-bridge`.

**What happens:**
- The `pum` WASM binary executes a GET request to `http://10.0.0.1/health`.
- In a real Apptron session, `10.0.0.1` is the gateway IP of the virtual network.
- The request is intercepted by the virtual network and forwarded to the Bridge Agent (which in mock mode is just a handler in the runner).
- The terminal should display: `Bridge Response: 200 OK`.

## Automated Tests
Run the Go test suite for the local components:
```bash
go test -v ./apptron/pkg/actions/... ./apptron/cmd/pum-admin/...
```
