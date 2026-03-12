#!/bin/bash
set -e

# Configuration
REPO_ROOT=$(git rev-parse --show-toplevel)
APPTRON_DIR="$REPO_ROOT/apptron"
BUILD_DIR="$REPO_ROOT/build/pum_bundle"
FINAL_ASSETS_DIR="$APPTRON_DIR/cmd/pum-admin/assets"
BUNDLE_DIR="$BUILD_DIR/bundle"
PUM_CLI_DIR="$APPTRON_DIR/services/pum-cli/cmd/pum-cli"

echo "Initializing build directory for pum bundle..."
rm -rf "$BUNDLE_DIR"
rm -rf "$BUILD_DIR/pum-bundle"
mkdir -p "$BUNDLE_DIR/bin"
mkdir -p "$FINAL_ASSETS_DIR/bundles"

echo "Compiling pum-cli to WASM..."
GOOS=js GOARCH=wasm go build -o "$BUNDLE_DIR/bin/pum.wasm" "$PUM_CLI_DIR/main.go" "$PUM_CLI_DIR/tui_wasm.go"

echo "Creating bin wrapper..."
cat <<'EOW' > "$BUNDLE_DIR/bin/pum"
#!/bin/sh
/bin/wexec /bin/pum.wasm "$@"
EOW
chmod +x "$BUNDLE_DIR/bin/pum"

echo "Customizing /etc/profile..."
mkdir -p "$BUNDLE_DIR/etc/profile.d"
cat <<'EOP' > "$BUNDLE_DIR/etc/profile"
export PATH=/bin:/usr/bin:/usr/local/bin
export PUM_SERVER_URL="https://api.pum-nms.local"
export PUM_MODE="mock"
echo "Welcome to the PUM Admin Distro (Phase 1: Mock Mode)"
EOP

echo "Packaging custom pum.tar.gz into runner assets..."

# Apptron's boot.go expects the contents to be inside a 'rootfs' directory within the tar
mkdir -p "$BUILD_DIR/pum-bundle/rootfs"
cp -r "$BUNDLE_DIR"/* "$BUILD_DIR/pum-bundle/rootfs/"
tar -C "$BUILD_DIR/pum-bundle" -czf "$FINAL_ASSETS_DIR/bundles/pum.tar.gz" rootfs

echo "Build complete! Custom PUM bundle is ready in $FINAL_ASSETS_DIR/bundles/pum.tar.gz"