#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

source "${ROOT_PATH}/infra/scripts/tasks/cache-env.sh"

cd "${ROOT_PATH}"

bash infra/scripts/tasks/cache-setup.sh

if [ ! -f pkg/hub/go.work ]; then
	cp pkg/hub/go.work.example pkg/hub/go.work
fi

(
	cd pkg/hub
	go work sync
)
