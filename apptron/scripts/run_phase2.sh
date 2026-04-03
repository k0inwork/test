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

echo "Re-packing sys.tar.gz using Python tarfile to guarantee precise ./ paths for Wanix..."
TMP_BUNDLE_DIR=$(mktemp -d)
tar -xzf "assets/bundles/sys.tar.gz" -C "$TMP_BUNDLE_DIR"
TARGET_TAR="$(pwd)/assets/bundles/sys.tar.gz"

cat << 'PY_TAR' > repack.py
import tarfile
import os

src = "$TMP_BUNDLE_DIR"
target = "$TARGET_TAR"

with tarfile.open(target, "w:gz") as tar:
    # Add root directory explicitly
    ti = tarfile.TarInfo(name="./")
    ti.type = tarfile.DIRTYPE
    ti.mode = 0o755
    tar.addfile(ti)

    for root, dirs, files in os.walk(src):
        for d in dirs:
            full_path = os.path.join(root, d)
            rel_path = os.path.relpath(full_path, src)
            arcname = "./" + rel_path + "/"
            tar.add(full_path, arcname=arcname, recursive=False)
        for f in files:
            full_path = os.path.join(root, f)
            rel_path = os.path.relpath(full_path, src)
            arcname = "./" + rel_path
            tar.add(full_path, arcname=arcname, recursive=False)
PY_TAR

sed -i "s|\$TMP_BUNDLE_DIR|$TMP_BUNDLE_DIR|g" repack.py
sed -i "s|\$TARGET_TAR|$TARGET_TAR|g" repack.py
python3 repack.py
rm repack.py
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
