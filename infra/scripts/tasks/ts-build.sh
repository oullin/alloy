#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

cd "${ROOT_PATH}"
pnpm exec turbo run build --filter='!@alloy/*-go'
