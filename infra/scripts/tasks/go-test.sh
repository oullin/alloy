#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

source "${ROOT_PATH}/infra/scripts/tasks/cache-env.sh"

API_PATH="${ROOT_PATH}/api"
GO_WORK_PATH="${API_PATH}/go.work"

# Every Go module under api owns a go.mod.
while IFS= read -r -d '' gomod; do
	module_dir="$(dirname "${gomod}")"
	echo "==> ${module_dir#"${ROOT_PATH}/"}"
	(
		cd "${module_dir}"
		export GOWORK="${GO_WORK_PATH}"
		if [[ "$(go env GOWORK)" != "${GO_WORK_PATH}" ]]; then
			echo "go workspace resolution is not using ${GO_WORK_PATH}" >&2
			exit 1
		fi
		go list -m github.com/oullin/alloy/api/contracts >/dev/null
		go vet ./...
		go test -race ./...
	)
done < <(find "${API_PATH}" -mindepth 2 -name go.mod -print0 | sort -z)
