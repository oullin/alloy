#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

source "${ROOT_PATH}/infra/scripts/tasks/cache-env.sh"

GO_PATH="${ROOT_PATH}/pkg/hub"
GO_WORK_PATH="${GO_PATH}/go.work"
GO_WORK_ENV="off"

if [[ -f "${GO_WORK_PATH}" ]]; then
	GO_WORK_ENV="${GO_WORK_PATH}"
fi

go_work_for_module() {
	local module_dir="$1"

	if [[ "${module_dir}" == "${GO_PATH}"* ]]; then
		printf '%s\n' "${GO_WORK_ENV}"
		return
	fi

	printf '%s\n' "off"
}

# Shared Go code lives under pkg/hub/. Web demo Go entrypoints live under web/*/api.
while IFS= read -r -d '' gomod; do
	module_dir="$(dirname "${gomod}")"
	module_work_env="$(go_work_for_module "${module_dir}")"
	echo "==> ${module_dir#"${ROOT_PATH}/"}"
	(
		cd "${module_dir}"
		export GOWORK="${module_work_env}"
		if [[ "${module_work_env}" != "off" && "$(go env GOWORK)" != "${module_work_env}" ]]; then
			echo "go workspace resolution is not using ${module_work_env}" >&2
			exit 1
		fi
		while IFS= read -r module_path; do
			if [[ "${module_path}" != "github.com/oullin/alloy/pkg/hub" && "${module_path}" != github.com/oullin/alloy/pkg/hub/* && "${module_path}" != alloy.dev/inertia-demo ]]; then
				echo "unexpected Go module path: ${module_path}" >&2
				exit 1
			fi
		done < <(go list -m)
		go vet ./...
		if [[ -n "${GO_COVERAGE_DIR:-}" ]]; then
			mkdir -p "${GO_COVERAGE_DIR}"
			module_name="${module_dir#"${ROOT_PATH}/"}"
			profile="${GO_COVERAGE_DIR}/$(echo "${module_name}" | tr '/' '-').out"
			go test -race -coverprofile="${profile}" ./...
			# go tool cover resolves sources through the current module, so it
			# must run here (inside the module) — not later from the repo root.
			total="$(go tool cover -func="${profile}" | tail -1 | awk '{print $NF}')"
			echo "${total}"
			printf '%s\t%s\n' "${module_name}" "${total}" >> "${GO_COVERAGE_DIR}/summary.tsv"
		else
			go test -race ./...
		fi
	)
done < <(
	find "${GO_PATH}" -name go.mod -print0
	find "${ROOT_PATH}/web" -path '*/api/go.mod' -print0 2>/dev/null || true
)
