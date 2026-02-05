#!/usr/bin/env bash
set -euo pipefail

# Build the Cierge reservation Lambda for AWS Lambda ARM64 runtime

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Building Cierge reservation Lambda..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap .

echo "Build complete: bootstrap ($(du -h bootstrap | cut -f1))"
