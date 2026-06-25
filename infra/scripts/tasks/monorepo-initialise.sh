#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

cd "${ROOT_PATH}"

if [ ! -f golang/go.work ]; then
	cp golang/go.work.example golang/go.work
fi

(
	cd golang
	go work sync
)
