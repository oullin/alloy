SHELL := /bin/bash

# Keep GO_IMAGE's Go version in sync with the `go` directive in the go.mod
# files (currently 1.26.5); that directive is the toolchain source of truth.
GO_IMAGE ?= golang:1.26.5-alpine
# Pinned go-task release. Bump deliberately (check github.com/go-task/task
# releases); overridable via `make TASK_VERSION=vX.Y.Z ...`.
TASK_VERSION ?= v3.52.0

export GO_IMAGE

VP := ./node_modules/.bin/vp
DOCKER_COMPOSE := $(shell if command -v docker-compose >/dev/null 2>&1; then printf 'docker-compose'; else printf 'docker compose'; fi)
DOCKER_RUN := $(DOCKER_COMPOSE) run --rm -T app
TASK := sh -lc 'apk add --no-cache bash git >/dev/null && git config --global --add safe.directory /workspace && export PATH=/go/bin:/usr/local/go/bin:$$PATH; go install github.com/go-task/task/v3/cmd/task@$(TASK_VERSION) && task "$$@"' task

export FMT_IMAGE ?= ghcr.io/oullin/go-fmt:v0.4.1-full
export PWD := $(shell pwd)

.DEFAULT_GOAL := help

# Vite+ owns orchestration. Make keeps existing commands working, including the
# Docker-backed formatter task defined in vite.config.ts.
.PHONY: help format format-all go-test go\:test FORCE
help:
	@$(VP) --help

format:
	@$(VP) run format

format-all:
	@$(VP) run format-all

go-test:
	@$(VP) run go:test

go\:test:
	@$(VP) run go:test

Makefile:
	@:

%: FORCE
	@$(VP) run $@

FORCE:
