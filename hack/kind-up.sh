#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CLUSTER="${CLUSTER:-salarythief}"
IMAGE="${IMAGE:-salarythief-collector:dev}"
MOCK_IMAGE="${MOCK_IMAGE:-salarythief-redfish-mock:dev}"

command -v kind >/dev/null || { echo "kind is required"; exit 1; }
command -v kubectl >/dev/null || { echo "kubectl is required"; exit 1; }
command -v docker >/dev/null || { echo "docker is required"; exit 1; }

if ! kind get clusters | grep -qx "$CLUSTER"; then
  kind create cluster --config "$ROOT/deploy/kind/cluster.yaml"
fi

docker build -t "$IMAGE" "$ROOT"
docker build -f "$ROOT/Dockerfile.mock" -t "$MOCK_IMAGE" "$ROOT"
kind load docker-image "$IMAGE" --name "$CLUSTER"
kind load docker-image "$MOCK_IMAGE" --name "$CLUSTER"

kubectl apply -f "$ROOT/deploy/kind/00-namespace.yaml"
kubectl apply -f "$ROOT/deploy/kind/10-redfish-mock.yaml"
kubectl apply -f "$ROOT/deploy/kind/20-collector.yaml"
kubectl apply -f "$ROOT/deploy/kind/30-opensearch.yaml"
kubectl apply -f "$ROOT/deploy/kind/40-prometheus.yaml"
kubectl apply -f "$ROOT/deploy/kind/50-grafana.yaml"
kubectl -n salarythief rollout restart deploy/salarythief-collector
kubectl -n observability create configmap grafana-dashboards \
  --from-file=salarythief-bmc.json="$ROOT/deploy/kind/grafana-dashboard.json" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Waiting for workloads..."
kubectl -n salarythief rollout status deploy/bmc-healthy-01 --timeout=180s
kubectl -n salarythief rollout status deploy/bmc-slow-01 --timeout=180s
kubectl -n salarythief rollout status deploy/bmc-partial-01 --timeout=180s
kubectl -n salarythief rollout status deploy/salarythief-collector --timeout=180s
kubectl -n observability rollout status deploy/prometheus --timeout=180s
kubectl -n observability rollout status deploy/opensearch --timeout=240s
kubectl -n observability rollout status deploy/grafana --timeout=180s
echo
echo "Local endpoints (kind extraPortMappings):"
echo "  Grafana     http://127.0.0.1:30000  (admin / admin)"
echo "  Prometheus  http://127.0.0.1:30090"
echo "  OpenSearch  http://127.0.0.1:30920"
echo "  Collector   kubectl -n salarythief port-forward svc/salarythief-collector 9100:9100"
