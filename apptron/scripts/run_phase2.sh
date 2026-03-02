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

# To properly and cleanly resolve missing components
if ! grep -q "assets/vscode.zip;" Makefile; then
    echo "Patching Makefile to download dependencies cleanly..."

    # 1. Update vscode extraction target to prevent curl errors and unzip natively
    # Use the local cache if available, else download cleanly.
    sed -i.bak 's/curl -sL $(VSCODE_URL) -o assets\/vscode.zip/if [ -f "..\/cmd\/pum-admin\/assets\/vscode.zip" ]; then cp "..\/cmd\/pum-admin\/assets\/vscode.zip" assets\/vscode.zip; else curl -sL $(VSCODE_URL) -o assets\/vscode.zip; fi/g' Makefile

    # 2. Prevent Docker from failing when downloading wanix by simply pulling from github release directly, instead of extracting docker blobs
    sed -i.bak '/$(DOCKER_CMD) rm -f apptron-wanix/d' Makefile
    sed -i.bak '/$(DOCKER_CMD) pull --platform linux\/amd64 ghcr.io\/tractordev\/wanix:runtime/d' Makefile
    sed -i.bak '/$(DOCKER_CMD) create --name apptron-wanix.*/d' Makefile
    sed -i.bak 's/$(DOCKER_CMD) cp apptron-wanix:\/wanix.min.js assets\/wanix.min.js/curl -sL https:\/\/github.com\/tractordev\/wanix\/releases\/download\/v0.3\/wanix_0.3_runtime.zip -o assets\/.tmp_wanix.zip \&\& unzip -p assets\/.tmp_wanix.zip assets\/wanix.min.js > assets\/wanix.min.js/g' Makefile
    sed -i.bak 's/$(DOCKER_CMD) cp apptron-wanix:\/wanix.js assets\/wanix.js/unzip -p assets\/.tmp_wanix.zip assets\/wanix.min.js > assets\/wanix.js \&\& rm -f assets\/.tmp_wanix.zip/g' Makefile
fi

make clean

# Ensure wrangler is installed locally in the worker package
echo "Installing worker dependencies..."
cd worker && npm ci && cd ..

make all

echo "Starting Phase 2 Worker in dev mode..."
echo "Mock Auth is ENABLED in apptron/worker/src/auth.ts."
echo "You can access the environment at http://localhost:8788"
cd worker && npx wrangler dev --port=8788
