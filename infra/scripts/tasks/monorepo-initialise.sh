#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

source "${ROOT_PATH}/infra/scripts/tasks/cache-env.sh"

cd "${ROOT_PATH}"

bash infra/scripts/tasks/cache-setup.sh

if [ ! -f api/go.work ]; then
	cp api/go.work.example api/go.work
fi

(
	cd api
	go work sync
)
