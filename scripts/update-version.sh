#!/bin/sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$ROOT_DIR"

if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  git rev-parse --short HEAD > .version
else
  echo "unknown" > .version
fi
