SHELL := /bin/bash

GO_IMAGE ?= golang:1.26.4-alpine
TASK_VERSION ?= latest
GO_BIN_VOLUME ?= alloy-go-bin
GO_MOD_VOLUME ?= alloy-go-mod
GO_BUILD_VOLUME ?= alloy-go-build

export GO_IMAGE
export GO_BIN_VOLUME
export GO_MOD_VOLUME
export GO_BUILD_VOLUME

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
