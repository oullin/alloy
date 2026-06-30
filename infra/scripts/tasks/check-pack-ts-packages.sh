#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:-}"

if [[ -z "${version}" ]]; then
	echo "VERSION is required" >&2
	exit 2
fi

artifact_dir="${RUNNER_TEMP:-/tmp}/alloy-ts-packages"
mkdir -p "${artifact_dir}"

packages=(
	"@alloy/console"
	"@alloy/money"
	"@alloy/navigator-routes"
	"@alloy/tempo"
	"@alloy/workflow"
)

for package in "${packages[@]}"; do
	package_version="$(pnpm --filter "${package}" exec node -p "require('./package.json').version")"

	if [[ "${package_version}" != "${version}" ]]; then
		echo "${package} version ${package_version} does not match release version ${version}" >&2
		exit 1
	fi

	pnpm --filter "${package}" typecheck
	pnpm --filter "${package}" test
	pnpm --filter "${package}" build
	pnpm --filter "${package}" pack --pack-destination "${artifact_dir}"
done

ls -lh "${artifact_dir}"
