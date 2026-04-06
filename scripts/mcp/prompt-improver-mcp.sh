#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

export GOWORK="${GOWORK:-off}"

cd "$REPO_ROOT"

if [[ -x "$REPO_ROOT/bin/prompt-improver" ]]; then
  exec "$REPO_ROOT/bin/prompt-improver" mcp "$@"
fi

exec go run . mcp "$@"
