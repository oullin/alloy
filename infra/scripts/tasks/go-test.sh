#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

cd "${ROOT_PATH}/packages/tempo-go"
GOWORK=off go test ./...
