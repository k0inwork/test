#!/bin/bash
set -e

REPO_ROOT=$(git rev-parse --show-toplevel)
APPTRON_DIR="$REPO_ROOT/apptron/apptron"

if [ ! -d "$APPTRON_DIR" ]; then
    echo "Apptron worker source not found. Make sure to fetch the submodule or clone apptron."
    return 1 2>/dev/null || false
fi

echo "Setting up Apptron worker..."
cd "$APPTRON_DIR"

if [ ! -f .env.local ]; then
    cp .env.example .env.local
fi

# Mock auth dynamically in the worker source
if ! grep -q "MOCK AUTH" worker/src/auth.ts; then
    echo "Mocking Auth in worker/src/auth.ts..."
    sed -i.bak '/export async function validateToken(/,/^}/c\
export async function validateToken(hankoApiUrl: string, token: string): Promise<boolean> {\
  // MOCK AUTH: Always return true\
  return true;\
}' worker/src/auth.ts
fi

# Always start from a clean Makefile to ensure patches apply correctly on multiple runs
git checkout Makefile 2>/dev/null || true

# To properly and cleanly resolve missing components
echo "Patching Makefile to download dependencies cleanly..."

# Use absolute path for copying to ensure it works reliably regardless of working directory
PUM_ADMIN_ASSETS="$REPO_ROOT/apptron/cmd/pum-admin/assets"

# 1. Update vscode extraction target to prevent curl errors and unzip natively
# Use the local cache if available, else download cleanly.
sed -i.bak "s|curl -sL \$(VSCODE_URL) -o assets/vscode.zip|if [ -f \"$PUM_ADMIN_ASSETS/vscode.zip\" ]; then cp \"$PUM_ADMIN_ASSETS/vscode.zip\" assets/vscode.zip; else curl -sL \$(VSCODE_URL) -o assets/vscode.zip; fi|g" Makefile

# 2. Prevent Docker from failing when downloading wanix by simply pulling from github release directly, instead of extracting docker blobs
sed -i.bak "s|\$(DOCKER_CMD) rm -f apptron-wanix|cp \"$PUM_ADMIN_ASSETS/wanix.min.js\" assets/wanix.min.js|g" Makefile
sed -i.bak "s|\$(DOCKER_CMD) pull --platform linux/amd64 ghcr.io/tractordev/wanix:runtime|cp \"$PUM_ADMIN_ASSETS/wanix.min.js\" assets/wanix.js|g" Makefile
sed -i.bak '/$(DOCKER_CMD) create --name apptron-wanix.*/d' Makefile
sed -i.bak '/$(DOCKER_CMD) cp apptron-wanix.*/d' Makefile

make clean

# Ensure wrangler is installed locally in the worker package
echo "Installing worker dependencies..."
cd worker && npm ci && cd ..

make all

echo "Starting Phase 2 Worker in dev mode..."
echo "Mock Auth is ENABLED in apptron/worker/src/auth.ts."
echo "You can access the environment at http://localhost:8788"
cd worker && npx wrangler dev --port=8788
