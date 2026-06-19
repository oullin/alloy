#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

cd "${ROOT_PATH}"

# Every Go module lives at packages/<domain>/<domain>-go and owns a go.mod.
while IFS= read -r -d '' gomod; do
	module_dir="$(dirname "${gomod}")"
	echo "==> ${module_dir#"${ROOT_PATH}/"}"
	(
		cd "${module_dir}"
		GOWORK=off go vet ./...
		GOWORK=off go test -race ./...
	)
done < <(find "${ROOT_PATH}/packages" -mindepth 3 -maxdepth 3 -name go.mod -print0 | sort -z)
