#!/bin/bash
# scripts/download_otelcol.sh
# Downloads the OpenTelemetry Collector Contrib executable

set -e

OTELCOL_VERSION="0.96.0"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

URL="https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v${OTELCOL_VERSION}/otelcol-contrib_${OTELCOL_VERSION}_${OS}_${ARCH}.tar.gz"

mkdir -p bin

if [ -f "bin/otelcol-contrib" ]; then
    echo "OpenTelemetry Collector is already downloaded."
    exit 0
fi

echo "Downloading OpenTelemetry Collector ${OTELCOL_VERSION} for ${OS}-${ARCH}..."
curl -sL -o otelcol.tar.gz "$URL"

tar -xzf otelcol.tar.gz
mv otelcol-contrib bin/

rm otelcol.tar.gz
# Also remove other root level files that tar might have extracted
rm -f README.md LICENSE

echo "OpenTelemetry Collector downloaded successfully to bin/otelcol-contrib"
