#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-}"
SCOPE="${2:-}"
ROOT_PATH="$(git rev-parse --show-toplevel)"

case "${MODE}" in
	changed)
		GIT_FLAGS=(--others --modified)
		;;
	all)
		GIT_FLAGS=(--cached --others)
		;;
	*)
		echo "Usage: bash infra/scripts/tasks/format-files.sh changed|all go|ts" >&2
		exit 2
		;;
esac

case "${SCOPE}" in
	go)
		FORMAT_GLOBS=('*.go')
		;;
	ts)
		FORMAT_GLOBS=('*.ts' '*.tsx' '*.vue' '*.mts' '*.cts')
		;;
	*)
		echo "Usage: bash infra/scripts/tasks/format-files.sh changed|all go|ts" >&2
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
	echo "No ${SCOPE} files to format."
	exit 0
fi

xargs -0 docker-compose run --rm -T \
	--user "$(id -u):$(id -g)" \
	fmt format < "${tmp}"
