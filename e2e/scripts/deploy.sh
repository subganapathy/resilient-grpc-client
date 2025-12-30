#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFESTS_DIR="$SCRIPT_DIR/../manifests"

echo "Deploying Prometheus..."
kubectl apply -f "$MANIFESTS_DIR/prometheus/prometheus-rbac.yaml"
kubectl apply -f "$MANIFESTS_DIR/prometheus/prometheus-config.yaml"
kubectl apply -f "$MANIFESTS_DIR/prometheus/prometheus-deployment.yaml"

echo "Waiting for Prometheus to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/prometheus

echo "Deploying test-server..."
kubectl apply -f "$MANIFESTS_DIR/apps/test-server.yaml"

echo "Waiting for test-server to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/test-server

echo "Deploying test-client..."
kubectl apply -f "$MANIFESTS_DIR/apps/test-client.yaml"

echo "Waiting for test-client to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/test-client

echo "Waiting for test-client to generate some metrics (30 seconds)..."
sleep 30

echo "Deployment complete!"
kubectl get pods
kubectl get services

