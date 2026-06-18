SHELL := /bin/bash

.DEFAULT_GOAL := help

# Task owns orchestration. This Makefile only keeps existing `make <target>`
# commands working.
.PHONY: help FORCE
help:
	@pnpm exec task --list

Makefile:
	@:

%: FORCE
	@pnpm exec task $@

FORCE:
