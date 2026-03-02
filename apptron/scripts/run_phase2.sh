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

# Copy wanix.js and wanix.min.js from pum-admin Phase 1 assets to avoid Docker downloads entirely
if [ ! -f "assets/wanix.js" ] && [ -f "$REPO_ROOT/apptron/cmd/pum-admin/assets/wanix.min.js" ]; then
    echo "Using existing wanix assets from pum-admin..."
    mkdir -p assets
    cp "$REPO_ROOT/apptron/cmd/pum-admin/assets/wanix.min.js" assets/wanix.min.js
    cp "$REPO_ROOT/apptron/cmd/pum-admin/assets/wanix.min.js" assets/wanix.js # wanix.js is not present there, so copy min.js or vice versa
fi

# Patch Makefile to use docker pull and save instead of docker create which fails in overlayfs within sandbox
if ! grep -q "wanix_temp/wanix.js" Makefile; then
    sed -i.bak 's/$(DOCKER_CMD) pull --platform linux\/amd64 ghcr.io\/tractordev\/wanix:runtime/$(DOCKER_CMD) pull ghcr.io\/tractordev\/wanix:runtime/g' Makefile
    sed -i.bak 's/$(DOCKER_CMD) create --name apptron-wanix.*/mkdir -p .\/wanix_temp \&\& $(DOCKER_CMD) save ghcr.io\/tractordev\/wanix:runtime -o .\/wanix_temp\/wanix.tar \&\& tar -xf .\/wanix_temp\/wanix.tar -C .\/wanix_temp\/ \&\& find .\/wanix_temp\/blobs\/sha256 -type f -exec sh -c '\''tar -xf "{}" -C .\/wanix_temp\/ 2>\/dev\/null'\'' \\;/g' Makefile
    sed -i.bak 's/$(DOCKER_CMD) cp apptron-wanix:\/wanix.min.js assets\/wanix.min.js/cp .\/wanix_temp\/wanix.min.js assets\/wanix.min.js/g' Makefile
    sed -i.bak 's/$(DOCKER_CMD) cp apptron-wanix:\/wanix.js assets\/wanix.js/cp .\/wanix_temp\/wanix.js assets\/wanix.js/g' Makefile
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
