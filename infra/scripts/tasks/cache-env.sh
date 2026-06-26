#!/usr/bin/env bash

ALLOY_CACHE_SCRIPT_PATH="${BASH_SOURCE[0]}"
ALLOY_ROOT_PATH="$(cd "$(dirname "${ALLOY_CACHE_SCRIPT_PATH}")/../../.." && pwd)"
ALLOY_CACHE_PATH="${ALLOY_ROOT_PATH}/infra/.cache"

mkdir -p \
	"${ALLOY_CACHE_PATH}/go/build" \
	"${ALLOY_CACHE_PATH}/go/mod" \
	"${ALLOY_CACHE_PATH}/go/path" \
	"${ALLOY_CACHE_PATH}/playwright" \
	"${ALLOY_CACHE_PATH}/pnpm/store" \
	"${ALLOY_CACHE_PATH}/tsbuild" \
	"${ALLOY_CACHE_PATH}/typescript" \
	"${ALLOY_CACHE_PATH}/vite" \
	"${ALLOY_CACHE_PATH}/vitest" \
	"${ALLOY_CACHE_PATH}/xdg"

export GOCACHE="${ALLOY_CACHE_PATH}/go/build"
export GOMODCACHE="${ALLOY_CACHE_PATH}/go/mod"
export GOPATH="${ALLOY_CACHE_PATH}/go/path"
export PLAYWRIGHT_BROWSERS_PATH="${ALLOY_CACHE_PATH}/playwright"
export XDG_CACHE_HOME="${ALLOY_CACHE_PATH}/xdg"
