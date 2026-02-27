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
mkdir -p "$BUNDLE_DIR/bin"
mkdir -p "$FINAL_ASSETS_DIR"

echo "Compiling pum-cli to WASM..."
# Use build tags to exclude TUI on WASM target
GOOS=js GOARCH=wasm go build -o "$BUNDLE_DIR/bin/pum.wasm" "$PUM_CLI_DIR/main.go" "$PUM_CLI_DIR/tui_wasm.go"

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

echo "Copying base Apptron assets to runner assets..."
cp -r "$APPTRON_SRC/assets/"* "$FINAL_ASSETS_DIR/"

echo "Packaging custom sys.tar.gz into runner assets..."
tar -C "$BUNDLE_DIR" -czf "$FINAL_ASSETS_DIR/sys.tar.gz" .

echo "Build complete! Custom Apptron assets are ready for embedding in $FINAL_ASSETS_DIR"
