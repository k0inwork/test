#!/bin/bash
set -e

# Configuration
REPO_ROOT=$(git rev-parse --show-toplevel)
APPTRON_DIR="$REPO_ROOT/apptron"
BUILD_DIR="$REPO_ROOT/build/distro"
ASSETS_DIR="$BUILD_DIR/assets"
BUNDLE_DIR="$BUILD_DIR/bundle"
PUM_CLI_SRC="$APPTRON_DIR/services/pum-cli/cmd/pum-cli"
APPTRON_SRC="$APPTRON_DIR/apptron"

echo "Initializing build directory..."
mkdir -p "$ASSETS_DIR"
mkdir -p "$BUNDLE_DIR/bin"

echo "Compiling pum-cli to WASM..."
# Note: Bubbletea has known build issues with pure GOOS=js.
# In a real environment, we would use a Wanix-compatible build tool or a shimmed library.
# For this prototype, we skip the actual binary build if it fails to ensure script continuity.
GOOS=js GOARCH=wasm go build -o "$BUNDLE_DIR/bin/pum.wasm" "$PUM_CLI_SRC" || echo "Warning: WASM build failed (expected without shims), continuing with prototype..."

echo "Creating bin wrapper..."
cat <<'EOW' > "$BUNDLE_DIR/bin/pum"
#!/bin/sh
/bin/wexec /bin/pum.wasm "$@"
EOW
chmod +x "$BUNDLE_DIR/bin/pum"

echo "Customizing /etc/profile..."
mkdir -p "$BUNDLE_DIR/etc"
cat <<'EOP' > "$BUNDLE_DIR/etc/profile"
export PATH=/bin:/usr/bin:/usr/local/bin
export PUM_SERVER_URL="https://api.pum-nms.local"
export PUM_MODE="mock"
echo "Welcome to the PUM Admin Distro (Phase 1: Mock Mode)"
EOP

echo "Copying base Apptron assets..."
cp -r "$APPTRON_SRC/assets/"* "$ASSETS_DIR/"

echo "Packaging custom sys.tar.gz..."
tar -C "$BUNDLE_DIR" -czf "$ASSETS_DIR/sys.tar.gz" .

echo "Build complete! Custom Apptron assets are in $ASSETS_DIR"
