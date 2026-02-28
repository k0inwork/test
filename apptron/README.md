# PUM Admin Center (Apptron Distro)

This directory contains the Apptron-based local-first management environment for the PUM-Go NMS.

## Quick Start (Phase 1: Mock Mode)

To build and run the Admin Center with mock data:

```bash
# From the repository root
bash apptron/scripts/run_phase1.sh
```

**What this script does:**
1.  Clones the Apptron source (if not already present).
2.  Compiles the `pum-cli` Go-WASM tool.
3.  Packages a custom `sys.tar.gz` filesystem bundle containing PUM tools.
4.  Compiles the `pum-admin` native Go runner.
5.  Starts a local server at `http://localhost:8080`.

## Building the Distro Manually

If you want to just build the assets without running the server:

```bash
bash apptron/scripts/build_distro.sh
```

The results will be in `build/distro/assets`.

## Directory Structure

- `apptron/`: The cloned Apptron source repository.
- `cmd/pum-admin/`: Native Go runner that serves assets and bridges the network.
- `services/pum-cli/`: The Go-WASM management tool (CLI/TUI).
- `pkg/actions/`: Shared logic for mirrored GUI/CLI actions.
- `docs/`: Detailed technical documentation.

## Testing

Run local Go tests:
```bash
go test -v ./apptron/pkg/actions/... ./apptron/cmd/pum-admin/...
```

Refer to `docs/testing.md` for interactive verification steps.
