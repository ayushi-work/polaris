#!/bin/bash
set -euo pipefail

CLUSTER_NAME="${1:-polaris}"

echo "Creating kind cluster: ${CLUSTER_NAME}"

if ! command -v kind &> /dev/null; then
    echo "Error: kind is not installed. Install it from: https://kind.sigs.k8s.io/"
    exit 1
fi

if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed."
    exit 1
fi

# Create cluster
kind create cluster --name "${CLUSTER_NAME}" --wait 120s

# Create namespace
kubectl create namespace polaris --dry-run=client -o yaml | kubectl apply -f -

# Build and load image
docker build -t polaris:latest -f deployments/Dockerfile .
kind load docker-image polaris:latest --name "${CLUSTER_NAME}"

# Apply RBAC
kubectl apply -f deployments/k8s/rbac.yaml -n polaris

# Apply deployment
kubectl apply -f deployments/k8s/deployment.yaml -n polaris

echo "Done! Cluster '${CLUSTER_NAME}' is ready."
echo "Run: kubectl port-forward -n polaris svc/polaris 8080:8080"
