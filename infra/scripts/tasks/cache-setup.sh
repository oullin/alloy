#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/cache-env.sh"

NODE_MODULES_PATH="${ALLOY_ROOT_PATH}/node_modules"
VITE_CACHE_LINK="${NODE_MODULES_PATH}/.vite"
VITE_TEMP_PATHS=(
	"${NODE_MODULES_PATH}/.vite-temp"
	"${ALLOY_ROOT_PATH}/infra/node_modules/.vite-temp"
	"${ALLOY_ROOT_PATH}/packages/console/node_modules/.vite-temp"
	"${ALLOY_ROOT_PATH}/packages/money/node_modules/.vite-temp"
	"${ALLOY_ROOT_PATH}/packages/tempo/node_modules/.vite-temp"
	"${ALLOY_ROOT_PATH}/packages/tempo/tests/node_modules/.vite-temp"
	"${ALLOY_ROOT_PATH}/packages/workflow/node_modules/.vite-temp"
)

mkdir -p "${NODE_MODULES_PATH}"
rm -rf "${VITE_CACHE_LINK}" "${VITE_TEMP_PATHS[@]}"
ln -s ../infra/.cache/vite "${VITE_CACHE_LINK}"

echo "Tool caches are configured under ${ALLOY_CACHE_PATH}."
