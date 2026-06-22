#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-}"
SCOPE="${2:-all}"
ROOT_PATH="$(git rev-parse --show-toplevel)"

case "${MODE}" in
	changed)
		GIT_FLAGS=(--others --modified)
		;;
	all)
		GIT_FLAGS=(--cached --others)
		;;
	*)
		echo "Usage: bash infra/scripts/tasks/format-files.sh changed|all [all|go|ts]" >&2
		exit 2
		;;
esac

case "${SCOPE}" in
	all)
		FORMAT_GLOBS=('*.go' '*.ts' '*.tsx' '*.vue' '*.mts' '*.cts')
		FORMAT_LABEL='Go and TypeScript'
		;;
	go)
		FORMAT_GLOBS=('*.go')
		FORMAT_LABEL='Go'
		;;
	ts)
		FORMAT_GLOBS=('*.ts' '*.tsx' '*.vue' '*.mts' '*.cts')
		FORMAT_LABEL='TypeScript'
		;;
	*)
		echo "Usage: bash infra/scripts/tasks/format-files.sh changed|all [all|go|ts]" >&2
		exit 2
		;;
esac

cd "${ROOT_PATH}"

tmp="$(mktemp)"
trap 'rm -f "${tmp}"' EXIT

git ls-files -z "${GIT_FLAGS[@]}" --exclude-standard -- "${FORMAT_GLOBS[@]}" |
	while IFS= read -r -d '' file; do
		if [ -f "${file}" ]; then
			printf '%s\0' "${file}"
		fi
	done > "${tmp}"

if [ ! -s "${tmp}" ]; then
	echo "No ${FORMAT_LABEL} files to format."
	exit 0
fi

xargs -0 docker-compose run --rm -T \
	--user "$(id -u):$(id -g)" \
	fmt format < "${tmp}"
