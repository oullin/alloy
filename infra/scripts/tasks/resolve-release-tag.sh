#!/usr/bin/env bash
set -euo pipefail

namespace="${RELEASE_NAMESPACE:-}"
version_pattern='^[0-9]+\.[0-9]+\.[0-9]+$'

case "${namespace}" in
	go | ts | treex)
		;;
	*)
		echo "RELEASE_NAMESPACE must be go, ts, or treex, got: ${namespace}" >&2
		exit 2
		;;
esac

if [[ -z "${GITHUB_OUTPUT:-}" ]]; then
	echo "GITHUB_OUTPUT is required" >&2
	exit 2
fi

if [[ "${GITHUB_EVENT_NAME:-}" == "workflow_dispatch" ]]; then
	version="${RELEASE_VERSION:-}"

	if [[ ! "${version}" =~ ${version_pattern} ]]; then
		echo "version must match X.Y.Z, got: ${version}" >&2
		exit 1
	fi

	tag="${namespace}/v${version}"

	if git ls-remote --exit-code --tags origin "refs/tags/${tag}" >/dev/null 2>&1; then
		echo "tag already exists: ${tag}" >&2
		exit 1
	fi
else
	tag="${GITHUB_REF_NAME:-}"
	version="${tag#"${namespace}/v"}"

	if [[ "${tag}" == "${version}" || ! "${version}" =~ ${version_pattern} ]]; then
		echo "expected a ${namespace}/vX.Y.Z tag, got: ${tag}" >&2
		exit 1
	fi
fi

echo "tag=${tag}" >> "${GITHUB_OUTPUT}"
echo "version=${version}" >> "${GITHUB_OUTPUT}"
