#!/usr/bin/env bash
# Full local verification (parity with GitHub Actions + recommended extras).
# Usage: ./scripts/verify.sh   OR   bash scripts/verify.sh [--race] [--skip-vuln] [--skip-lint]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

RACE=0
SKIP_VULN=0
SKIP_LINT=0
for arg in "$@"; do
  case "$arg" in
    --race) RACE=1 ;;
    --skip-vuln) SKIP_VULN=1 ;;
    --skip-lint) SKIP_LINT=1 ;;
  esac
done

echo "Repository: $ROOT"
echo ""

echo "=== go version ==="
go version

echo ""
echo "=== go mod verify ==="
go mod verify

if [[ "$SKIP_LINT" -eq 0 ]] && command -v golangci-lint >/dev/null 2>&1; then
  echo ""
  echo "=== golangci-lint run ==="
  golangci-lint run --timeout=5m
elif [[ "$SKIP_LINT" -eq 0 ]]; then
  echo ""
  echo "=== golangci-lint (skipped: not in PATH) ==="
fi

echo ""
echo "=== go vet ./... ==="
go vet ./...

if [[ "$SKIP_VULN" -eq 0 ]]; then
  echo ""
  echo "=== govulncheck ./... ==="
  go install golang.org/x/vuln/cmd/govulncheck@latest
  govulncheck ./...
fi

echo ""
echo "=== go build ./... ==="
go build -v ./...

echo ""
echo "=== go build cmd (explicit) ==="
go build -v ./cmd/uqda
go build -v ./cmd/uqdactl

echo ""
if [[ "$RACE" -eq 1 ]]; then
  echo "=== go test -race ./... (slower) ==="
  go test ./... -count=1 -race -timeout=120s
else
  echo "=== go test ./... ==="
  go test ./... -count=1 -timeout=120s
fi

echo ""
echo "All checks passed."
