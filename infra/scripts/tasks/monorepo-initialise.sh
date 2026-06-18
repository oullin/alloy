#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="$(git rev-parse --show-toplevel)"

cd "${ROOT_PATH}"
cp go.work.example go.work
go work sync
