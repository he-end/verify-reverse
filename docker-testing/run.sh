#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

cleanup() {
    echo ""
    echo "Stopping Docker containers..."
    docker compose --project-directory "$SCRIPT_DIR" down
}

trap cleanup EXIT

echo "Building images and starting services..."
docker compose --project-directory "$SCRIPT_DIR" up --build
