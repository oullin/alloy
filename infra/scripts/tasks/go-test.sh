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

# Test one Go module. Output goes to per-invocation files so parallel modules
# never interleave their logs or coverage summaries. Coverage rows are written to
# a private per-module file and concatenated in module order by the caller.
run_module() {
	local module_dir="$1"
	local module_work_env="$2"
	local summary_file="$3"

	cd "${module_dir}"
	export GOWORK="${module_work_env}"
	if [[ "${module_work_env}" != "off" && "$(go env GOWORK)" != "${module_work_env}" ]]; then
		echo "go workspace resolution is not using ${module_work_env}" >&2
		return 1
	fi
	while IFS= read -r module_path; do
		if [[ "${module_path}" != "github.com/oullin/alloy/pkg/hub" && "${module_path}" != github.com/oullin/alloy/pkg/hub/* && "${module_path}" != alloy.dev/inertia-demo ]]; then
			echo "unexpected Go module path: ${module_path}" >&2
			return 1
		fi
	done < <(go list -m)
	# `go test` re-runs `go vet` by default. Keep one explicit vet pass so vet
	# failures are attributable, and pass `-vet=off` to `go test` to drop the
	# duplicate analysis pass.
	go vet ./...
	if [[ -n "${GO_COVERAGE_DIR:-}" ]]; then
		mkdir -p "${GO_COVERAGE_DIR}"
		local module_name profile total
		module_name="${module_dir#"${ROOT_PATH}/"}"
		profile="${GO_COVERAGE_DIR}/$(echo "${module_name}" | tr '/' '-').out"
		go test -vet=off -race -coverprofile="${profile}" ./...
		# go tool cover resolves sources through the current module, so it
		# must run here (inside the module) — not later from the repo root.
		total="$(go tool cover -func="${profile}" | tail -1 | awk '{print $NF}')"
		echo "${total}"
		printf '%s\t%s\n' "${module_name}" "${total}" > "${summary_file}"
	else
		go test -vet=off -race ./...
	fi
}

# Shared Go code lives under pkg/hub/. Web demo Go entrypoints live under web/*/api.
# Modules are independent, so run them concurrently and aggregate results after.
work_dir="$(mktemp -d)"
trap 'rm -rf "${work_dir}"' EXIT

pids=()
module_labels=()
log_files=()
summary_files=()

index=0
while IFS= read -r -d '' gomod; do
	module_dir="$(dirname "${gomod}")"
	module_work_env="$(go_work_for_module "${module_dir}")"
	module_labels[index]="${module_dir#"${ROOT_PATH}/"}"
	log_files[index]="${work_dir}/module-${index}.log"
	summary_files[index]="${work_dir}/module-${index}.tsv"
	run_module "${module_dir}" "${module_work_env}" "${summary_files[index]}" \
		> "${log_files[index]}" 2>&1 &
	pids[index]=$!
	index=$((index + 1))
done < <(
	find "${GO_PATH}" -name go.mod -print0
	find "${ROOT_PATH}/web" -path '*/api/go.mod' -print0 2>/dev/null || true
)

status=0
for i in "${!pids[@]}"; do
	if ! wait "${pids[i]}"; then
		status=1
	fi
	echo "==> ${module_labels[i]}"
	cat "${log_files[i]}"
done

# Concatenate per-module coverage rows in module order so summary.tsv is stable
# and never interleaved by concurrent writers.
if [[ -n "${GO_COVERAGE_DIR:-}" ]]; then
	mkdir -p "${GO_COVERAGE_DIR}"
	: > "${GO_COVERAGE_DIR}/summary.tsv"
	for i in "${!summary_files[@]}"; do
		if [[ -s "${summary_files[i]}" ]]; then
			cat "${summary_files[i]}" >> "${GO_COVERAGE_DIR}/summary.tsv"
		fi
	done
fi

exit "${status}"
