#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

cd "${ROOT_PATH}"

if [ ! -f api/go.work ]; then
	cp api/go.work.example api/go.work
fi

(
	cd api
	go work sync
)
