#!/bin/bash

# Setup Python virtual environment
VENV_DIR=".venv"
if [ ! -d "$VENV_DIR" ]; then
    echo "Creating Python virtual environment..."
    python3 -m venv "$VENV_DIR"
fi
source "$VENV_DIR/bin/activate"

# Install pyyaml
python3 -m pip install pyyaml

BIN_DIR=".bin"
mkdir -p "$BIN_DIR"

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

if [ "$OS" = "Linux" ]; then
    ARCH_KEY="linux_url"
elif [ "$OS" = "Darwin" ]; then
    if [ "$ARCH" = "arm64" ]; then
        ARCH_KEY="darwin_arm64_url"
    else
        ARCH_KEY="darwin_amd64_url"
    fi
else
    echo "Unsupported OS: $OS"
    # return 1
fi

download_dependency() {
    local DEP_NAME=$1
    local BIN_NAME=$2
    local URL=$(python3 -c "import yaml; d = yaml.safe_load(open('deploy_config.yaml')); print(d['external_dependencies']['$DEP_NAME']['$ARCH_KEY'])")

    if [ ! -f "$BIN_DIR/$BIN_NAME" ]; then
        echo "Downloading $DEP_NAME from $URL..."
        wget -qO- "$URL" | tar xvz -C "$BIN_DIR"

        find "$BIN_DIR" -name "$BIN_NAME" -type f -exec mv {} "$BIN_DIR/" \;
        chmod +x "$BIN_DIR/$BIN_NAME"

        if [ "$DEP_NAME" = "prometheus" ]; then
            find "$BIN_DIR" -name "prometheus.yml" -type f -exec mv {} ./prometheus.yml \;
        fi
    fi
}

echo "Ensuring Goreman is installed..."
if ! command -v goreman &> /dev/null; then
    go install github.com/mattn/goreman@latest
    export PATH="$PATH:$(go env GOPATH)/bin"
fi

echo "Checking external dependencies..."
download_dependency "jaeger" "jaeger-all-in-one"
download_dependency "prometheus" "prometheus"
download_dependency "otelcol" "otelcol-contrib"

echo "Generating environment files..."
python3 scripts/generate_env.py

echo "Starting services with Goreman..."
if command -v goreman &> /dev/null; then
    goreman start
else
    $(go env GOPATH)/bin/goreman start
fi
