#!/bin/bash

set -e

CLUSTER_NAME="rgrpc-e2e"

echo "Creating kind cluster: $CLUSTER_NAME"

# Check if cluster already exists
if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    echo "Cluster $CLUSTER_NAME already exists. Deleting it first..."
    kind delete cluster --name "$CLUSTER_NAME"
fi

# Create kind cluster
kind create cluster --name "$CLUSTER_NAME" --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 8080
    protocol: TCP
  - containerPort: 443
    hostPort: 8443
    protocol: TCP
EOF

echo "Waiting for cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=120s --context kind-${CLUSTER_NAME}

echo "Kind cluster $CLUSTER_NAME is ready!"
kubectl cluster-info --context kind-${CLUSTER_NAME}

