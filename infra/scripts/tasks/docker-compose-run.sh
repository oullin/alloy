#!/usr/bin/env bash
set -euo pipefail

if command -v docker-compose >/dev/null 2>&1; then
	exec docker-compose run --rm -T "$@"
fi

exec docker compose run --rm -T "$@"
