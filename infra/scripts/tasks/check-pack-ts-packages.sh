#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:-}"

if [[ -z "${version}" ]]; then
	echo "VERSION is required" >&2
	exit 2
fi

artifact_dir="${RUNNER_TEMP:-/tmp}/alloy-ts-packages"
rm -rf "${artifact_dir}"
mkdir -p "${artifact_dir}"

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cd "${repo_root}"

# Discovered from the workspace rather than hardcoded, so a sixth package cannot
# silently skip this gate by never being added to a list. Read with a while loop
# rather than mapfile: macOS ships bash 3.2, so this has to run locally too.
packages=()

while IFS= read -r line; do
	[[ -n "${line}" ]] && packages+=("${line}")
done < <(pnpm --filter './sdk/*' list --depth -1 --json | node -e '
	let raw = "";
	process.stdin.on("data", (chunk) => (raw += chunk));
	process.stdin.on("end", () => {
		for (const entry of JSON.parse(raw)) {
			if (entry.name) console.log(entry.name);
		}
	});
' | sort)

if [[ "${#packages[@]}" -eq 0 ]]; then
	echo "no packages discovered under sdk/*" >&2
	exit 1
fi

echo "Releasing ${#packages[@]} packages at ${version}:"
printf '  %s\n' "${packages[@]}"

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

echo
echo "Validating packed manifests..."

shopt -s nullglob
tarballs=("${artifact_dir}"/*.tgz)
shopt -u nullglob

if [[ "${#tarballs[@]}" -ne "${#packages[@]}" ]]; then
	echo "expected ${#packages[@]} tarballs, found ${#tarballs[@]}" >&2
	exit 1
fi

# Consumers pin the release artifact by URL, so the filename is part of the
# contract and a 404 is how they find out it moved. The count check above only
# proves five files exist, not that each package produced the name its URL
# implies -- so assert the name derived from the package name and version.
#
# This catches a change in how pnpm names artifacts, and a package that packed
# under an unexpected name. It cannot catch a deliberate rename, since that
# moves both sides together; that remains a coordinated change with consumers.
for package in "${packages[@]}"; do
	expected="$(printf '%s' "${package}" | sed 's|^@||; s|/|-|g')-${version}.tgz"

	if [[ ! -f "${artifact_dir}/${expected}" ]]; then
		echo "${package} did not produce ${expected}; downstream URLs pin that name" >&2
		exit 1
	fi
done

for tarball in "${tarballs[@]}"; do
	node infra/scripts/tasks/check-packed-manifest.mjs "${tarball}" "${version}"
done

echo
echo "Running publint and are-the-types-wrong..."

report_dir="$(mktemp -d)"
consumer_dir="$(mktemp -d)"
trap 'rm -rf "${report_dir}" "${consumer_dir}"' EXIT

for tarball in "${tarballs[@]}"; do
	pnpm --silent dlx publint "${tarball}"

	# Redirect rather than pipe: are-the-types-wrong exits without draining
	# stdout, so a pipe truncates the report at ~128KB. Keep it out of the
	# artifact directory so only tarballs are published.
	attw_report="${report_dir}/$(basename "${tarball}").attw.json"
	# `|| true`: attw exits non-zero whenever any problem exists, including the
	# node10/cjs ones that are expected for ESM-only packages. The node check
	# below decides what actually blocks a release.
	pnpm --silent dlx @arethetypeswrong/cli "${tarball}" --format json >"${attw_report}" || true
	node infra/scripts/tasks/check-packed-types.mjs "${attw_report}" "$(basename "${tarball}")"
done

echo
echo "Verifying a clean external install..."

(
	cd "${consumer_dir}"
	npm init -y >/dev/null
	# Installs from the tarballs alone, with no workspace and no registry, which
	# is the closest offline approximation of what a customer receives.
	npm install --no-audit --no-fund "${tarballs[@]}" >/dev/null

	for package in "${packages[@]}"; do
		node --input-type=module -e "await import('${package}'); console.log('  ok import ${package}')"
	done
)

echo
ls -lh "${artifact_dir}"
