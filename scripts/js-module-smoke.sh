#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd)"

cd "$ROOT"

go test ./pkg/services/smailnailjs ./pkg/js/modules/smailnail -count=1

echo "smailnail JS module smoke passed."
