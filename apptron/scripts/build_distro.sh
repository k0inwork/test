#!/bin/bash
set -e

# Configuration
REPO_ROOT=$(git rev-parse --show-toplevel)
APPTRON_DIR="$REPO_ROOT/apptron"
BUILD_DIR="$REPO_ROOT/build/distro"
FINAL_ASSETS_DIR="$APPTRON_DIR/cmd/pum-admin/assets"
BUNDLE_DIR="$BUILD_DIR/bundle"
PUM_CLI_DIR="$APPTRON_DIR/services/pum-cli/cmd/pum-cli"
APPTRON_SRC="$APPTRON_DIR/apptron"

echo "Initializing build directory..."
rm -rf "$BUNDLE_DIR"
mkdir -p "$BUNDLE_DIR/bin"
mkdir -p "$FINAL_ASSETS_DIR"

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


echo "Packaging custom sys.tar.gz into runner assets..."
mkdir -p "$FINAL_ASSETS_DIR/bundles"

# Apptron's boot.go expects the contents to be inside a 'rootfs' directory within the tar
mkdir -p "$BUILD_DIR/sys-bundle/rootfs"
cp -r "$BUNDLE_DIR"/* "$BUILD_DIR/sys-bundle/rootfs/"

# Add v86 and bios files to sys-bundle
rm -rf "$BUILD_DIR/sys-bundle/v86"
mkdir -p "$BUILD_DIR/sys-bundle/v86"

# Get files from assets if available
if [ -d "$APPTRON_DIR/scripts/v86" ]; then
    cp "$APPTRON_DIR/scripts/v86/"* "$BUILD_DIR/sys-bundle/v86/" 2>/dev/null || true
fi

# Add kernel to sys-bundle
rm -rf "$BUILD_DIR/sys-bundle/kernel"
mkdir -p "$BUILD_DIR/sys-bundle/kernel"
if [ -d "$APPTRON_DIR/scripts/kernel" ]; then
    cp "$APPTRON_DIR/scripts/kernel/"* "$BUILD_DIR/sys-bundle/kernel/" 2>/dev/null || true
fi

tar -C "$BUILD_DIR/sys-bundle" -czf "$FINAL_ASSETS_DIR/bundles/sys.tar.gz" rootfs v86 kernel

echo "Build complete! Custom Apptron assets are ready for embedding in $FINAL_ASSETS_DIR"
