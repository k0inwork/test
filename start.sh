#!/bin/bash

set -e

BIN_DIR=".bin"
mkdir -p "$BIN_DIR"

# Setup Python virtual environment to avoid global pip install issues on macOS
VENV_DIR=".venv"
if [ ! -d "$VENV_DIR" ]; then
    echo "Creating Python virtual environment..."
    python3 -m venv "$VENV_DIR"
fi
source "$VENV_DIR/bin/activate"

# Install pyyaml if not present
if ! python3 -c "import yaml" >/dev/null 2>&1; then
    echo "Installing pyyaml for Python generator script..."
    python3 -m pip install pyyaml
fi

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

get_arch_key() {
    if [ "$OS" = "Linux" ]; then
        echo "linux_url"
    elif [ "$OS" = "Darwin" ]; then
        if [ "$ARCH" = "arm64" ]; then
            echo "darwin_arm64_url"
        else
            echo "darwin_amd64_url"
        fi
    else
        echo "Unsupported OS: $OS"
        exit 1
    fi
}

ARCH_KEY=$(get_arch_key)

download_dependency() {
    local DEP_NAME=$1
    local BIN_NAME=$2
    local URL=$(python3 -c "import yaml; d = yaml.safe_load(open('deploy_config.yaml')); print(d['external_dependencies']['$DEP_NAME']['$ARCH_KEY'])")

    if [ ! -f "$BIN_DIR/$BIN_NAME" ]; then
        echo "Downloading $DEP_NAME from $URL..."
        wget -qO- "$URL" | tar xvz -C "$BIN_DIR"

        # Flatten extraction (move the binary out of its versioned folder directly into .bin)
        find "$BIN_DIR" -name "$BIN_NAME" -type f -exec mv {} "$BIN_DIR/" \;
        chmod +x "$BIN_DIR/$BIN_NAME"
    fi
}

echo "Ensuring Goreman is installed..."
if ! command -v goreman &> /dev/null; then
    echo "Goreman not found in PATH, installing via go install..."
    go install github.com/mattn/goreman@latest
    export PATH="$PATH:$(go env GOPATH)/bin"
fi

echo "Checking external dependencies..."
download_dependency "jaeger" "jaeger-all-in-one"
download_dependency "prometheus" "prometheus"

echo "Generating environment files..."
python3 scripts/generate_env.py

echo "Starting services with Goreman..."
if command -v goreman &> /dev/null; then
    goreman start
else
    $(go env GOPATH)/bin/goreman start
fi
