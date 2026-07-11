#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

cleanup() {
    echo ""
    echo "Cleaning up — stopping postgres container..."
    docker compose --project-directory "$ROOT_DIR" down
}

trap cleanup EXIT

echo "Starting postgres container..."
docker compose --project-directory "$ROOT_DIR" up -d --wait

echo ""
echo "Running tests..."
go test -count=1 -timeout 60s ./verify/... "$@"

echo ""
echo "All tests passed"
