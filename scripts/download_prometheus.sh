#!/bin/bash
# scripts/download_prometheus.sh
# Downloads the Prometheus executable if not already present

set -e

PROMETHEUS_VERSION="2.51.0"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

URL="https://github.com/prometheus/prometheus/releases/download/v${PROMETHEUS_VERSION}/prometheus-${PROMETHEUS_VERSION}.${OS}-${ARCH}.tar.gz"

mkdir -p bin

if [ -f "bin/prometheus" ]; then
    echo "Prometheus is already downloaded."
    exit 0
fi

echo "Downloading Prometheus ${PROMETHEUS_VERSION} for ${OS}-${ARCH}..."
curl -sL -o prometheus.tar.gz "$URL"

tar -xzf prometheus.tar.gz
mv "prometheus-${PROMETHEUS_VERSION}.${OS}-${ARCH}/prometheus" bin/
mv "prometheus-${PROMETHEUS_VERSION}.${OS}-${ARCH}/promtool" bin/

rm -rf "prometheus-${PROMETHEUS_VERSION}.${OS}-${ARCH}"
rm prometheus.tar.gz

echo "Prometheus downloaded successfully to bin/prometheus"
