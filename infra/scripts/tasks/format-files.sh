#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-}"
ROOT_PATH="$(git rev-parse --show-toplevel)"
FMT_IMAGE="ghcr.io/oullin/go-fmt:v0.4.0"
FORMAT_GLOBS=('*.go' '*.ts' '*.tsx' '*.vue' '*.mts' '*.cts')

case "${MODE}" in
	changed)
		GIT_FLAGS=(--others --modified)
		;;
	all)
		GIT_FLAGS=(--cached --others)
		;;
	*)
		echo "Usage: bash infra/scripts/tasks/format-files.sh changed|all" >&2
		exit 2
		;;
esac

cd "${ROOT_PATH}"

tmp="$(mktemp)"
trap 'rm -f "${tmp}"' EXIT

git ls-files -z "${GIT_FLAGS[@]}" --exclude-standard -- "${FORMAT_GLOBS[@]}" |
	while IFS= read -r -d '' file; do
		[ -f "${file}" ] && printf '%s\0' "${file}"
	done > "${tmp}"

if [ ! -s "${tmp}" ]; then
	echo "No Go/TypeScript files to format."
	exit 0
fi

xargs -0 docker run --rm \
	--user "$(id -u):$(id -g)" \
	-v "${ROOT_PATH}:/work" \
	-w /work \
	-e HOST_PROJECT_PATH="${ROOT_PATH}" \
	"${FMT_IMAGE}" format < "${tmp}"
