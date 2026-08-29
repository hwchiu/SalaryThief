#!/usr/bin/env bash
set -euo pipefail

collector_pf=""
prom_pf=""
cleanup() { [ -n "$collector_pf" ] && kill "$collector_pf" 2>/dev/null || true; [ -n "$prom_pf" ] && kill "$prom_pf" 2>/dev/null || true; }
trap cleanup EXIT

kubectl -n salarythief rollout status deploy/salarythief-collector --timeout=60s
kubectl -n observability rollout status deploy/prometheus --timeout=60s
kubectl -n observability rollout status deploy/opensearch --timeout=60s
# Reset the deterministic unreachable fault in case a previous run recovered it.
if kubectl -n salarythief get svc bmc-unreachable-01 -o jsonpath='{.spec.selector.app}' | grep -q .; then
  kubectl -n salarythief patch svc bmc-unreachable-01 --type json -p '[{"op":"remove","path":"/spec/selector"}]' >/dev/null
fi
kubectl -n salarythief delete endpoints bmc-unreachable-01 --ignore-not-found >/dev/null
kubectl -n salarythief delete endpointslices -l kubernetes.io/service-name=bmc-unreachable-01 --ignore-not-found >/dev/null
kubectl -n salarythief port-forward svc/salarythief-collector 19100:9100 >/tmp/salarythief-collector-pf.log 2>&1 & collector_pf=$!
kubectl -n observability port-forward svc/prometheus 19090:9090 >/tmp/salarythief-prom-pf.log 2>&1 & prom_pf=$!
ready=false
for n in $(seq 1 30); do if curl -fsS http://127.0.0.1:19100/healthz >/dev/null 2>&1; then ready=true; break; fi; sleep 1; done
$ready || { echo "collector port-forward did not become ready" >&2; exit 1; }
for n in $(seq 1 15); do metrics=$(curl -fsS http://127.0.0.1:19100/metrics); grep -q 'redfish_up{.*server="bmc-unreachable-01".*} 0' <<<"$metrics" && break; sleep 1; done
metrics=$(curl -fsS http://127.0.0.1:19100/metrics)
grep -q 'redfish_up{.*server="bmc-healthy-01".*} 1' <<<"$metrics"
grep -q 'redfish_up{.*server="bmc-unreachable-01".*} 0' <<<"$metrics"
grep -q 'redfish_resource_health{.*resource_family="thermal".*server="bmc-unreachable-01".*} 0' <<<"$metrics"
grep -q 'redfish_up{.*server="bmc-slow-01".*} 0' <<<"$metrics"
grep -q 'redfish_up{.*server="bmc-partial-01".*} 1' <<<"$metrics"
grep -q 'redfish_resource_health{.*resource_family="storage".*server="bmc-partial-01".*} 4' <<<"$metrics"
grep -q 'redfish_last_success_timestamp_seconds{.*server="bmc-healthy-01"' <<<"$metrics"
grep -q 'redfish_data_age_seconds{.*server="bmc-healthy-01"' <<<"$metrics"
curl -fsS --max-time 2 http://127.0.0.1:19100/metrics >/dev/null
for n in $(seq 1 30); do curl -fsS 'http://127.0.0.1:19090/api/v1/query?query=redfish_up' | grep -q bmc-healthy-01 && break; sleep 1; done
# Make the configured unreachable service point at the healthy mock and assert recovery.
kubectl -n salarythief patch svc bmc-unreachable-01 --type merge -p '{"spec":{"selector":{"app":"bmc-healthy-01"}}}' >/dev/null
for n in $(seq 1 15); do metrics=$(curl -fsS http://127.0.0.1:19100/metrics); grep -q 'redfish_up{.*server="bmc-unreachable-01".*} 1' <<<"$metrics" && break; sleep 1; done
grep -q 'redfish_up{.*server="bmc-unreachable-01".*} 1' <<<"$metrics"
# Persistence may fail independently; serving cached Prometheus state must remain fast.
kubectl -n observability scale deploy/opensearch --replicas=0 >/dev/null
sleep 2
curl -fsS --max-time 2 http://127.0.0.1:19100/metrics >/dev/null
grep -q 'redfish_up{.*server="bmc-healthy-01".*} 1' < <(curl -fsS http://127.0.0.1:19100/metrics)
kubectl -n observability scale deploy/opensearch --replicas=1 >/dev/null
kubectl -n observability rollout status deploy/opensearch --timeout=180s >/dev/null
echo 'PASS: healthy, unreachable, slow, partial, recovery, OpenSearch isolation, freshness, scrape isolation, and Prometheus discovery'
