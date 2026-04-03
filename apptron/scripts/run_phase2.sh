#!/bin/bash
export PUM_ENV="development"
set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../.." && pwd)
APPTRON_DIR="$REPO_ROOT/apptron/apptron"

if [ ! -d "$APPTRON_DIR" ]; then
    echo "Apptron worker source not found. Cloning tractordev/apptron..."
    git clone --depth 1 https://github.com/tractordev/apptron "$APPTRON_DIR"
fi

echo "Setting up Apptron worker..."
cd "$APPTRON_DIR"

if [ ! -f .env.local ]; then
    cp .env.example .env.local
fi

# Run the patcher scripts
echo "Applying Apptron source patches..."
python3 "$SCRIPT_DIR/patch_apptron.py"


# Always start from a clean Makefile to ensure patches apply correctly on multiple runs
git checkout Makefile 2>/dev/null || true
git checkout Dockerfile 2>/dev/null || true

# Run the build patcher script
echo "Applying Apptron build patches..."
export PUM_ADMIN_ASSETS="$REPO_ROOT/apptron/cmd/pum-admin/assets"
python3 "$SCRIPT_DIR/patch_build.py"


# Copy the checked-in base sys.tar.gz bundle from the repo to the worker assets BEFORE make all
echo "Ensuring base sys.tar.gz is present from repository..."
mkdir -p assets/bundles
if [ ! -f "assets/bundles/sys.tar.gz" ]; then
    cp "$REPO_ROOT/apptron/assets/bundles/sys.tar.gz" "assets/bundles/sys.tar.gz"
fi

# Only run make clean and make all if assets/wanix.wasm is missing
if [ ! -f "assets/wanix.wasm" ]; then
    echo "Base Apptron assets not found. Running make clean and make all..."
    make clean
    # Ensure wrangler is installed locally in the worker package
    echo "Installing worker dependencies..."
    cd worker && npm ci && cd ..

    # make clean deletes sys.tar.gz, so copy it back before make all runs
    mkdir -p assets/bundles
    cp "$REPO_ROOT/apptron/assets/bundles/sys.tar.gz" "assets/bundles/sys.tar.gz"

    make all
else
    echo "Base Apptron assets found. Skipping make all to speed up startup."
    echo "Installing worker dependencies..."
    cd worker && npm ci && cd ..
fi

# Check if CLI needs rebuilding by comparing timestamps
CLI_CHANGED=false
if [ ! -f "$REPO_ROOT/build/distro/bundle/bin/pum.wasm" ]; then
    CLI_CHANGED=true
elif [ -n "$(find "$REPO_ROOT/apptron/services/pum-cli" -newer "$REPO_ROOT/build/distro/bundle/bin/pum.wasm" 2>/dev/null | head -n 1)" ]; then
    CLI_CHANGED=true
fi


if [ "$CLI_CHANGED" = true ] || [ ! -f "assets/bundles/pum.tar.gz" ]; then
    echo "pum-cli source changed or pum.tar.gz missing, building..."
    (cd "$REPO_ROOT" && bash "apptron/scripts/build_distro.sh")
fi

echo "Re-packing sys.tar.gz using standard tar to guarantee precise ./ paths for Wanix..."
TMP_BUNDLE_DIR=$(mktemp -d)
tar -xzf "assets/bundles/sys.tar.gz" -C "$TMP_BUNDLE_DIR"
TARGET_TAR="$(pwd)/assets/bundles/sys.tar.gz"
(cd "$TMP_BUNDLE_DIR" && tar -czf "$TARGET_TAR" .)
rm -rf "$TMP_BUNDLE_DIR"

echo "Copying sys.tar.gz and pum.tar.gz into worker assets..."

mkdir -p assets/bundles
cp "$REPO_ROOT/apptron/assets/bundles/sys.tar.gz" "assets/bundles/sys.tar.gz"
cp "$REPO_ROOT/apptron/cmd/pum-admin/assets/bundles/pum.tar.gz" "assets/bundles/pum.tar.gz"

echo "Starting Phase 2 Worker in dev mode..."
echo "Mock Auth is ENABLED in apptron/worker/src/auth.ts."
echo "You can access the environment at http://localhost:8788"


echo "Clearing Wrangler cache..."
rm -rf .wrangler
docker image prune -f --filter "label=org.opencontainers.image.title=worker"

export CI=true
export WRANGLER_SEND_METRICS=false
cd worker && npx wrangler dev --port=8788 --log-level=none
