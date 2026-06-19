#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

cd "${ROOT_PATH}"
if [ -f go.work ]; then
	echo "go.work already exists; refusing to overwrite." >&2
	exit 1
fi

cp go.work.example go.work
go work sync
