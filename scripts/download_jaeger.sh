#!/bin/bash
# scripts/download_jaeger.sh
# Downloads the Jaeger all-in-one executable if not already present

set -e

JAEGER_VERSION="1.54.0"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

if [ "$OS" = "darwin" ] && [ "$ARCH" = "amd64" ]; then
    # Jaeger stopped publishing macos amd64 in some recent builds or uses unified ones. Assuming standard structure
    OS="darwin"
fi

URL="https://github.com/jaegertracing/jaeger/releases/download/v${JAEGER_VERSION}/jaeger-${JAEGER_VERSION}-${OS}-${ARCH}.tar.gz"

mkdir -p bin

if [ -f "bin/jaeger-all-in-one" ]; then
    echo "Jaeger is already downloaded."
    exit 0
fi

echo "Downloading Jaeger ${JAEGER_VERSION} for ${OS}-${ARCH}..."
curl -sL -o jaeger.tar.gz "$URL"

tar -xzf jaeger.tar.gz
mv "jaeger-${JAEGER_VERSION}-${OS}-${ARCH}/jaeger-all-in-one" bin/

rm -rf "jaeger-${JAEGER_VERSION}-${OS}-${ARCH}"
rm jaeger.tar.gz

echo "Jaeger downloaded successfully to bin/jaeger-all-in-one"
