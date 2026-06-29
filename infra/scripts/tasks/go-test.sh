#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

source "${ROOT_PATH}/infra/scripts/tasks/cache-env.sh"

GO_PATH="${ROOT_PATH}/go"
GO_WORK_PATH="${GO_PATH}/go.work"
GO_WORK_ENV="off"

if [[ -f "${GO_WORK_PATH}" ]]; then
	GO_WORK_ENV="${GO_WORK_PATH}"
fi

# Every Go module under go owns a go.mod.
while IFS= read -r -d '' gomod; do
	module_dir="$(dirname "${gomod}")"
	echo "==> ${module_dir#"${ROOT_PATH}/"}"
	(
		cd "${module_dir}"
		export GOWORK="${GO_WORK_ENV}"
		if [[ "${GO_WORK_ENV}" != "off" && "$(go env GOWORK)" != "${GO_WORK_ENV}" ]]; then
			echo "go workspace resolution is not using ${GO_WORK_ENV}" >&2
			exit 1
		fi
		module_path="$(go list -m)"
		if [[ "${module_path}" != "alloy.dev/go" && "${module_path}" != alloy.dev/go/* ]]; then
			echo "unexpected Go module path: ${module_path}" >&2
			exit 1
		fi
		go vet ./...
		go test -race ./...
	)
done < <(find "${GO_PATH}" -name go.mod -print0 | sort -z)
