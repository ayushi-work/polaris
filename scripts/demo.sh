#!/bin/bash
set -euo pipefail

echo "=== Polaris Demo ==="
echo ""

echo "1. Starting API server in dry-run mode..."
go run ./cmd/polaris serve --dry-run &
SERVER_PID=$!
sleep 2

cleanup() {
    kill $SERVER_PID 2>/dev/null || true
}
trap cleanup EXIT

echo "2. Creating test incidents..."
curl -s -X POST http://localhost:8080/api/v1/incidents \
  -H "Content-Type: application/json" \
  -d '{"namespace":"default","kind":"Pod","resource_name":"payment-api-7d4f8","incident_type":"OOMKilled","severity":"critical","message":"Container exceeded memory limit"}'
echo ""

curl -s -X POST http://localhost:8080/api/v1/incidents \
  -H "Content-Type: application/json" \
  -d '{"namespace":"production","kind":"Pod","resource_name":"checkout-worker-3b2a1","incident_type":"CrashLoopBackOff","severity":"critical","message":"Application crash loop detected"}'
echo ""

curl -s -X POST http://localhost:8080/api/v1/incidents \
  -H "Content-Type: application/json" \
  -d '{"namespace":"default","kind":"Pod","resource_name":"frontend-9c8x2","incident_type":"ImagePullBackOff","severity":"warning","message":"Cannot pull image from registry"}'
echo ""

echo "3. Listing incidents..."
curl -s http://localhost:8080/api/v1/incidents | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8080/api/v1/incidents
echo ""

echo "4. Creating chaos scenario..."
curl -s -X POST http://localhost:8080/api/v1/chaos/scenarios \
  -H "Content-Type: application/json" \
  -d '{"name":"delete-payment-pod","description":"Delete a random payment pod to test resilience","action":"delete_pod"}'
echo ""

echo ""
echo "Demo complete!"
