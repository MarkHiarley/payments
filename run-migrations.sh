#!/usr/bin/env bash
# Kept for backwards compatibility: this now delegates to `make migrate`.
# Prefer calling `make migrate` / `make migrate-down` directly.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if [ ! -f .env ]; then
    echo "No .env found — creating it from .env.example."
    cp -n .env.example .env
fi

echo "Applying migrations with 'make migrate'..."
make migrate
