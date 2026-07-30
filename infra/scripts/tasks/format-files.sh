#!/usr/bin/env bash
set -euo pipefail

# Thin wrapper over fmtkit (https://github.com/oullin/fmtkit), which formats both
# Go and TS/Vue from a single binary. It replaces the previous Docker-backed
# formatter: no image to pull, no daemon, and nothing for CI runners to install
# beyond the binary itself.
#
# fmtkit does its own changed-file discovery (vs HEAD, plus untracked) and
# honours .gitignore/.prettierignore, so this script only maps the repo's
# changed|all + all|go|ts interface onto its subcommands.
#
# The ${arr[@]+"${arr[@]}"} expansions below are deliberate: macOS ships bash
# 3.2, where "${empty[@]}" under `set -u` is an unbound-variable error.

MODE="${1:-}"
SCOPE="${2:-all}"

usage() {
	echo "Usage: bash infra/scripts/tasks/format-files.sh changed|all [all|go|ts]" >&2
}

if ! command -v fmtkit >/dev/null 2>&1; then
	cat >&2 <<-'MSG'
		fmtkit is not installed.

		  brew install --cask fmtkit

		Or download a release binary from https://github.com/oullin/fmtkit.
	MSG

	exit 127
fi

case "${SCOPE}" in
	all)
		scope_flags=()
		;;
	go)
		scope_flags=(--go)
		;;
	ts)
		scope_flags=(--ts)
		;;
	*)
		usage
		exit 2
		;;
esac

cd "$(git rev-parse --show-toplevel)"

case "${MODE}" in
	changed)
		exec fmtkit format ${scope_flags[@]+"${scope_flags[@]}"} .
		;;
	all)
		exec fmtkit format-all ${scope_flags[@]+"${scope_flags[@]}"}
		;;
	*)
		usage
		exit 2
		;;
esac
