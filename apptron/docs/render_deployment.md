# Deploying the PUM Admin Center to Render

While the PUM Admin Center is designed to run locally, you can also host a central instance of the Admin Distro on Render for shared access or demonstrations.

## Render Deployment Strategy

### 1. Web Service Deployment
The `pum-admin` Unified Runner can be deployed as a **Render Web Service**.
- **Runtime**: Go
- **Build Command**: `bash apptron/scripts/build_distro.sh && go build -o build/pum-admin apptron/cmd/pum-admin/main.go`
- **Start Command**: `./build/pum-admin`
- **Environment Variables**:
  - `PUM_MODE`: `mock` (for public demos) or `live` (for internal use).
  - `PORT`: `8080` (Render automatically routes traffic to the `PORT` env var).

### 2. Microservices Deployment
You can deploy the PUM-Go microservices (Identity, Product, etc.) as separate Render services or in a monorepo configuration.

### 3. Database & Cache
- **Render Postgres**: Use for persistent storage of nodes, users, and configuration templates.
- **Render Key-Value (Redis)**: Use for real-time task tracking and state management.

## Using Render MCP Tools
As your engineer, I can use the Render MCP to:
- **Monitor Deploys**: Track the status of your builds and deployments.
- **View Logs**: Debug the `pum-admin` server or microservices in real-time.
- **Check Metrics**: Monitor CPU/Memory usage of your network management components.
- **Manage Infrastructure**: Scale services or update environment variables directly.
