# Building the PUM Admin Distro

This document describes how to build a customized Apptron distribution that comes pre-loaded with the PUM-Go Admin Center tools.

## Distro Components

The "PUM Admin Distro" consists of:
1.  **Customized Alpine Bundle (`sys.tar.gz`)**: Contains the `pum-cli` binary and pre-configured environment scripts.
2.  **PUM Dashboards**: Integrated HTML/JS dashboards served as part of the Apptron assets.
3.  **Pre-configured Workspace**: VSCode layout optimized for network administration and PUM scripting.
4.  **WASM Headless Core**: The background service that powers the dashboards.

## Build Process

### 1. Compile WASM Binaries
Compile all PUM Go tools to WebAssembly:
```bash
GOOS=js GOARCH=wasm go build -o build/pum-cli.wasm ./services/pum-cli/cmd/pum-cli
```

### 2. Create the Filesystem Bundle
The bundle is a compressed tarball that Apptron loads at boot.
1.  Extract the base Apptron `sys.tar.gz`.
2.  Copy `pum-cli.wasm` to `/bin/pum`.
3.  Add a custom shell alias or wrapper script to `/bin/wexec` for `pum`.
4.  Add PUM environment variables to `/etc/profile`.
5.  Repack the bundle.

### 3. Integrate Assets
1.  Copy PUM web dashboards to the `assets/pum/` directory.
2.  Update `assets/index.html` to show PUM-specific project templates or a "PUM Command Center" shortcut.
3.  Inject the WASM Headless Core (`pum-core.wasm`) into the assets to be loaded by a WebWorker.

## Deployment
The resulting `assets/` directory (including the custom `sys.tar.gz`) can be hosted on any static site provider or deployed to Cloudflare Workers using the Apptron worker logic.

## Build Script Prototype
See `scripts/build_distro.sh` for an automated implementation of this process.
