#!/bin/bash
set -e

REPO_ROOT=$(git rev-parse --show-toplevel)
APPTRON_DIR="$REPO_ROOT/apptron"
APPTRON_SRC="$APPTRON_DIR/apptron"

if [ ! -d "$APPTRON_SRC" ]; then
    echo "Apptron source not found. Cloning..."
    git clone --depth 1 https://github.com/tractordev/apptron "$APPTRON_SRC"
fi

echo "Building PUM Admin Distro (Phase 1)..."
bash "$APPTRON_DIR/scripts/build_distro.sh"

echo "Building Unified Runner..."
go build -o "$REPO_ROOT/build/pum-admin" "$APPTRON_DIR/cmd/pum-admin/main.go"

echo "Starting PUM Admin Center in MOCK mode..."
export PUM_MODE="mock"
export PUM_ADMIN_PORT="8080"

echo "--------------------------------------------------"
echo "Admin Center starting at http://localhost:8080"
echo "Verify health at: curl http://localhost:8080/health"
echo "Once in the browser terminal, try: pum ping-bridge"
echo "Press Ctrl+C to stop."
echo "--------------------------------------------------"

"$REPO_ROOT/build/pum-admin"
