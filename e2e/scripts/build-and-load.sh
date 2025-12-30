#!/bin/bash

set -e

CLUSTER_NAME="rgrpc-e2e"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "Building and loading images into kind cluster: $CLUSTER_NAME"

# Build test-server image (from root to access proto package)
echo "Building test-server..."
cd "$ROOT_DIR"
docker build -f e2e/test-server/Dockerfile -t test-server:latest .

# Build test-client image (from root to access rgrpc and proto packages)
echo "Building test-client..."
cd "$ROOT_DIR"
docker build -f e2e/test-client/Dockerfile -t test-client:latest .

# Load images into kind
echo "Loading images into kind cluster..."
kind load docker-image test-server:latest --name "$CLUSTER_NAME"
kind load docker-image test-client:latest --name "$CLUSTER_NAME"

echo "Images loaded successfully!"

