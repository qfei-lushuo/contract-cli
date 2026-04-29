#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"

grep -q "go build -trimpath" "$ROOT_DIR/build.sh"
grep -q "go build -trimpath" "$ROOT_DIR/scripts/build-release-assets.sh"
grep -q "go build -trimpath" "$ROOT_DIR/tests/cli_e2e/smoke.sh"
grep -q "go build" "$ROOT_DIR/tests/release/local-install.sh"
grep -q -- "-trimpath" "$ROOT_DIR/tests/release/local-install.sh"
grep -q "go install -trimpath" "$ROOT_DIR/Makefile"

echo "build flags ok"
