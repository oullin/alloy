SHELL := /bin/bash

GO_IMAGE ?= golang:1.26.4-alpine
GO_BIN_VOLUME ?= alloy-go-bin
GO_MOD_VOLUME ?= alloy-go-mod
GO_BUILD_VOLUME ?= alloy-go-build

export GO_IMAGE
export GO_BIN_VOLUME
export GO_MOD_VOLUME
export GO_BUILD_VOLUME

VP := ./node_modules/.bin/vp

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
