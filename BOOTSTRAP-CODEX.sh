#!/usr/bin/env bash
set -euo pipefail

echo "SalaryThief Codex bootstrap"
echo

if [ ! -d .git ]; then
  echo "ERROR: run this script from the SalaryThief repository root."
  exit 1
fi

echo "Repository:"
git remote -v | head -n 2 || true
echo

echo "Branch:"
git branch --show-current
echo

echo "Working tree:"
git status --short
echo

echo "Tool versions:"
(go version || true)
(docker version --format '{{.Client.Version}}' || docker --version || true)
(kind version || true)
(kubectl version --client || true)
echo

echo "Relevant files:"
find internal -maxdepth 3 -type f 2>/dev/null | sort || true
find deploy/kind -maxdepth 2 -type f 2>/dev/null | sort || true
echo

echo "Baseline unit tests:"
if go test ./...; then
  echo "BASELINE_GO_TEST=PASS"
else
  echo "BASELINE_GO_TEST=FAIL"
fi

echo
echo "Next: start Codex in this repo and paste the prompt from START-CODEX.md"
