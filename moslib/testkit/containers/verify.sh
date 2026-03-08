#!/usr/bin/env bash
set -euo pipefail

echo "=== Container simulation verification ==="

LOGS_DIR="${1:-.}"

check_output() {
  local svc="$1"
  local file="$LOGS_DIR/$svc.log"
  if [ ! -f "$file" ]; then
    echo "FAIL: missing output for $svc"
    return 1
  fi
  if ! grep -q '"status":"ok"' "$file"; then
    echo "FAIL: $svc did not report ok"
    cat "$file"
    return 1
  fi
  echo "PASS: $svc completed successfully"
}

check_output "ide-a1"
check_output "ide-a2"
check_output "ide-b1"
check_output "ide-b2"

echo "=== All container scenarios passed ==="
