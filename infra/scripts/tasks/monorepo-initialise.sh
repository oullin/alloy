#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

source "${ROOT_PATH}/infra/scripts/tasks/cache-env.sh"

cd "${ROOT_PATH}"

bash infra/scripts/tasks/cache-setup.sh

if [ ! -f go/go.work ]; then
	cp go/go.work.example go/go.work
fi

(
	cd go
	go work sync
)
