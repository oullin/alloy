#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

cd "${ROOT_PATH}"
bash infra/scripts/tasks/docker-compose-run.sh fmt check
