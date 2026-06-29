SHELL := /bin/bash

VP := ./node_modules/.bin/vp

export FMT_IMAGE ?= ghcr.io/oullin/go-fmt:v0.4.1-full
export PWD := $(shell pwd)

.DEFAULT_GOAL := help

# Vite+ owns orchestration. Make keeps existing commands working, including the
# Docker-backed formatter task defined in vite.config.ts.
.PHONY: help format format-all backend-test backend\:test go-test go\:test FORCE
help:
	@$(VP) --help

format:
	@$(VP) run format

format-all:
	@$(VP) run format-all

backend-test:
	@$(VP) run backend:test

backend\:test:
	@$(VP) run backend:test

go-test:
	@$(VP) run go:test

go\:test:
	@$(VP) run go:test

Makefile:
	@:

%: FORCE
	@$(VP) run $@

FORCE:
