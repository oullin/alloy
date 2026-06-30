#!/usr/bin/env bash
set -euo pipefail

base_sha="${DETECT_BASE_SHA:-}"
head_sha="${DETECT_HEAD_SHA:-}"
required_engine="${DETECT_REQUIRED_ENGINE:-}"
allow_unchanged="${DETECT_ALLOW_UNCHANGED:-false}"

if [[ -z "${base_sha}" || -z "${head_sha}" ]]; then
	echo "DETECT_BASE_SHA and DETECT_HEAD_SHA are required" >&2
	exit 2
fi

case "${required_engine}" in
	"" | go | ts)
		;;
	*)
		echo "DETECT_REQUIRED_ENGINE must be go or ts, got: ${required_engine}" >&2
		exit 2
		;;
esac

case "${allow_unchanged}" in
	true | false)
		;;
	*)
		echo "DETECT_ALLOW_UNCHANGED must be true or false, got: ${allow_unchanged}" >&2
		exit 2
		;;
esac

if [[ -z "${GITHUB_OUTPUT:-}" ]]; then
	echo "GITHUB_OUTPUT is required" >&2
	exit 2
fi

if [[ -z "${GITHUB_STEP_SUMMARY:-}" ]]; then
	echo "GITHUB_STEP_SUMMARY is required" >&2
	exit 2
fi

changed_files="$(mktemp "${RUNNER_TEMP:-/tmp}/alloy-changed-files.XXXXXX")"
trap 'rm -f "${changed_files}"' EXIT

git diff --name-only "${base_sha}" "${head_sha}" > "${changed_files}"

labels=""
if [[ -n "${GITHUB_EVENT_PATH:-}" && -f "${GITHUB_EVENT_PATH}" ]]; then
	labels="$(jq -r '.pull_request.labels[].name' "${GITHUB_EVENT_PATH}")"
fi

has_coverage_label=false
has_e2e_label=false

if grep -Fxq 'coverage-all' <<< "${labels}"; then
	has_coverage_label=true
fi

if grep -Eq '^e2e-' <<< "${labels}"; then
	has_e2e_label=true
fi

has_special_label=false
if [[ "${has_coverage_label}" == "true" || "${has_e2e_label}" == "true" ]]; then
	has_special_label=true
fi

go_changed=false
ts_changed=false
browser_changed=false

while IFS= read -r file; do
	[[ -n "${file}" ]] || continue

	case "${file}" in
		go/* | go.work | go.work.example | .github/actions/setup/* | .github/workflows/ci.yml | .github/workflows/ci-go.yml | .github/workflows/release-go.yml | package.json | pnpm-lock.yaml | pnpm-workspace.yaml | vite.config.ts | .npmrc | infra/scripts/tasks/detect-changed-surfaces.sh | infra/scripts/tasks/go-test.sh | infra/scripts/tasks/cache-env.sh)
			go_changed=true
			;;
	esac

	case "${file}" in
		ts/* | infra/* | packages/* | web/* | .github/actions/setup/* | .github/workflows/ci.yml | .github/workflows/ci-ts.yml | .github/workflows/release-ts.yml | package.json | pnpm-lock.yaml | pnpm-workspace.yaml | vite.config.ts | tsconfig.json | .npmrc)
			ts_changed=true
			;;
	esac
done < "${changed_files}"

required_engine_changed=true
case "${required_engine}" in
	go)
		required_engine_changed="${go_changed}"
		;;
	ts)
		required_engine_changed="${ts_changed}"
		;;
esac

{
	echo "go_changed=${go_changed}"
	echo "ts_changed=${ts_changed}"
	echo "browser_changed=${browser_changed}"
	echo "has_special_label=${has_special_label}"
	echo "has_coverage_label=${has_coverage_label}"
	echo "has_e2e_label=${has_e2e_label}"
	echo "required_engine_changed=${required_engine_changed}"
	echo "allow_unchanged=${allow_unchanged}"
} >> "${GITHUB_OUTPUT}"

{
	echo "### Engine selection"
	echo "- Go changed: ${go_changed}"
	echo "- TypeScript changed: ${ts_changed}"
	echo "- Browser/E2E changed: ${browser_changed}"
	echo "- coverage-all label: ${has_coverage_label}"
	echo "- e2e-* label: ${has_e2e_label}"
	if [[ -n "${required_engine}" ]]; then
		echo "- Required engine: ${required_engine}"
		echo "- Required engine changed: ${required_engine_changed}"
		echo "- Allow unchanged release: ${allow_unchanged}"
	fi
} >> "${GITHUB_STEP_SUMMARY}"

if [[ "${required_engine_changed}" != "true" && "${allow_unchanged}" != "true" ]]; then
	echo "no ${required_engine} changes detected between ${base_sha} and ${head_sha}" >&2
	exit 1
fi
