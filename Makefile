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

DOCKER_RUN := docker-compose run --rm -T app
TASK := sh -lc 'apk add --no-cache bash git >/dev/null && git config --global --add safe.directory /workspace && export PATH=/go/bin:/usr/local/go/bin:$$PATH; go install github.com/go-task/task/v3/cmd/task@$(TASK_VERSION) && task "$$@"' task

export FMT_IMAGE ?= ghcr.io/oullin/go-fmt:v0.4.1-full
export PWD := $(shell pwd)

.DEFAULT_GOAL := help

# Task owns orchestration. Make keeps existing commands working and mirrors the
# Dockerised DockTUI path for Go and formatting tasks.
.PHONY: help format format-all go-test go\:test FORCE
help:
	@pnpm exec task --list

format:
	@vp run format

format-all:
	@vp run format-all

go-test:
	$(DOCKER_RUN) $(TASK) go:test

go\:test:
	$(DOCKER_RUN) $(TASK) go:test

Makefile:
	@:

%: FORCE
	@pnpm exec task $@

FORCE:
