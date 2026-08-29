#!/usr/bin/env bash
set -euo pipefail
CLUSTER="${CLUSTER:-salarythief}"
kind delete cluster --name "$CLUSTER"
