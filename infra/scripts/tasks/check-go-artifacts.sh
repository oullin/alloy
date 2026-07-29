#!/usr/bin/env bash
set -euo pipefail

# Builds the Go module artifacts and proves an external consumer can resolve
# them, using only a local file:// proxy. This runs before anything is uploaded,
# so a bad zip layout or an unstripped replace directive fails here rather than
# in a customer's build.

version="${VERSION:-}"

if [[ -z "${version}" ]]; then
	echo "VERSION is required (e.g. VERSION=v0.3.0)" >&2
	exit 2
fi

if [[ "${version}" != v* ]]; then
	version="v${version}"
fi

repo_root="$(git rev-parse --show-toplevel)"
artifact_dir="${GO_ARTIFACT_DIR:-${RUNNER_TEMP:-/tmp}/alloy-go-modules}"

rm -rf "${artifact_dir}"
mkdir -p "${artifact_dir}"

consumer_dir="$(mktemp -d)"
trap 'rm -rf "${consumer_dir}"' EXIT

echo "Building module artifacts at ${version}..."

# The tag's commit time, so .info reflects the release rather than the build.
commit_time="$(git -C "${repo_root}" show -s --format=%cI HEAD)"

(
	cd "${repo_root}/infra/modzip"
	GOWORK=off go run . \
		-repo "${repo_root}" \
		-version "${version}" \
		-out "${artifact_dir}" \
		-time "${commit_time}"
)

echo
echo "Verifying an external consumer can resolve them..."

# One representative package per published module. Importing them proves the
# zip carries real, compilable source and not just a well-formed shell.
probes=(
	"hara.sh/alloy/str"
	"hara.sh/alloy/auth/passkeys"
	"hara.sh/alloy/queue/drivers/sqs/credentials"
)

modules=(
	"hara.sh/alloy"
	"hara.sh/alloy/auth/passkeys"
	"hara.sh/alloy/queue/drivers/sqs"
)

(
	cd "${consumer_dir}"
	go mod init alloy.dev/artifact-check >/dev/null

	# GONOSUMDB, deliberately not GOPRIVATE: GOPRIVATE implies GONOPROXY, which
	# would make the go command ignore this file:// proxy and try to reach the
	# real hara.sh from CI.
	export GOWORK=off
	export GOFLAGS=-mod=mod
	export GONOSUMDB="hara.sh/alloy"
	export GOPROXY="file://${artifact_dir},https://proxy.golang.org,direct"

	for module in "${modules[@]}"; do
		go get "${module}@${version}"
		echo "  ok resolved ${module}@${version}"
	done

	{
		echo "package main"
		echo
		echo "import ("
		for probe in "${probes[@]}"; do
			echo "	_ \"${probe}\""
		done
		echo ")"
		echo
		echo "func main() {}"
	} >main.go

	go mod tidy >/dev/null
	go build ./...

	echo "  ok built a consumer importing all three modules"
	echo
	echo "Consumer go.sum entries:"
	grep 'hara.sh/alloy' go.sum | sed 's/^/  /'
)

echo
echo "Artifacts in ${artifact_dir}:"
find "${artifact_dir}" -type f | sed "s|${artifact_dir}|  .|" | sort
