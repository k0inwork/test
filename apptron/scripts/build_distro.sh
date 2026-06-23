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
rm -rf "$BUILD_DIR/sys-bundle"
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

echo "Installing wanix CLI if missing..."
mkdir -p "$REPO_ROOT/build/bin"
if [ ! -f "$REPO_ROOT/build/bin/wanix" ]; then
    echo "Building wanix CLI..."
    rm -rf /tmp/wanix-build
    git clone https://github.com/tractordev/wanix /tmp/wanix-build
    (cd /tmp/wanix-build && git checkout 44f753f37865a842f5471f2874994deef30ee83d)

    # Patch wanix CLI to handle symlinks correctly to avoid 'archive/tar: write too long'
    cat << 'EOF' > /tmp/wanix-build/patch_wanix.py
import sys
with open("cmd/wanix/bundle.go", "r") as f:
    c = f.read()
c = c.replace('header, err := tar.FileInfoHeader(info, "")', '''link := ""
\t\t\t\tif info.Mode()&os.ModeSymlink != 0 {
\t\t\t\t\tlink, _ = os.Readlink(path)
\t\t\t\t}
\t\t\t\theader, err := tar.FileInfoHeader(info, link)''')
c = c.replace("if !info.IsDir() {", "if info.Mode().IsRegular() {")
with open("cmd/wanix/bundle.go", "w") as f:
    f.write(c)
EOF
    (cd /tmp/wanix-build && python3 patch_wanix.py)

    (cd /tmp/wanix-build && go build -o "$REPO_ROOT/build/bin/wanix" ./cmd/wanix)
fi

echo "Creating pum.tar.gz bundle via wanix bundle pack..."
"$REPO_ROOT/build/bin/wanix" bundle pack "$BUNDLE_DIR" "$FINAL_ASSETS_DIR/bundles/pum.tar.gz"

echo "Build complete! Custom Apptron assets are ready for embedding in $FINAL_ASSETS_DIR"
