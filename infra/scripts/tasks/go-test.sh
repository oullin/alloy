#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

API_PATH="${ROOT_PATH}/api"

# Every Go module under api owns a go.mod.
while IFS= read -r -d '' gomod; do
	module_dir="$(dirname "${gomod}")"
	echo "==> ${module_dir#"${ROOT_PATH}/"}"
	(
		cd "${module_dir}"
		GOWORK=off go vet ./...
		GOWORK=off go test -race ./...
	)
done < <(find "${API_PATH}" -mindepth 2 -name go.mod -print0 | sort -z)
