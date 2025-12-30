#!/bin/bash

set -e

CLUSTER_NAME="rgrpc-e2e"

echo "Cleaning up e2e test resources..."

# Delete deployments
kubectl delete deployment test-client test-server prometheus --ignore-not-found=true
kubectl delete service test-server prometheus --ignore-not-found=true
kubectl delete configmap prometheus-config --ignore-not-found=true

echo "Deployments cleaned up."

read -p "Do you want to delete the kind cluster? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Deleting kind cluster: $CLUSTER_NAME"
    kind delete cluster --name "$CLUSTER_NAME"
    echo "Cluster deleted."
else
    echo "Kind cluster preserved. Delete manually with: kind delete cluster --name $CLUSTER_NAME"
fi

