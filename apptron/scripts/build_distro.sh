#!/bin/bash
set -e

# Configuration
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../.." && pwd)
APPTRON_DIR="$REPO_ROOT/apptron"
BUILD_DIR="$REPO_ROOT/build/distro"
FINAL_ASSETS_DIR="$APPTRON_DIR/cmd/pum-admin/assets"
BUNDLE_DIR="$BUILD_DIR/bundle"
PUM_CLI_DIR="$APPTRON_DIR/services/pum-cli/cmd/pum-cli"
APPTRON_SRC="$APPTRON_DIR/apptron"

echo "Initializing build directory..."

rm -rf "$BUNDLE_DIR"
mkdir "$BUNDLE_DIR"
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


echo "Packaging custom pum.tar.gz into runner assets..."
mkdir -p "$FINAL_ASSETS_DIR/bundles"

# Apptron's boot.go expects the contents to be inside a 'rootfs' directory within the tar



echo "Creating pum.tar.gz bundle..."




# Apptron expects contents to be inside a 'rootfs' directory within the tar
mkdir -p "$BUNDLE_DIR/rootfs/bin"
mkdir -p "$BUNDLE_DIR/rootfs/etc/profile.d"
mv "$BUNDLE_DIR/bin/"* "$BUNDLE_DIR/rootfs/bin/"
mv "$BUNDLE_DIR/etc/"* "$BUNDLE_DIR/rootfs/etc/"
rmdir "$BUNDLE_DIR/bin" "$BUNDLE_DIR/etc"

(cd "$BUNDLE_DIR" && tar -czf "$FINAL_ASSETS_DIR/bundles/pum.tar.gz" .)


echo "Build complete! Custom Apptron assets are ready for embedding in $FINAL_ASSETS_DIR"
