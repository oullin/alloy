SHELL := /bin/bash

PROVISION_DIR := provision

.DEFAULT_GOAL := help

.PHONY: help FORCE
help:
	@$(MAKE) --no-print-directory -C $(PROVISION_DIR) help

Makefile:
	@:

%: FORCE
	@$(MAKE) --no-print-directory -C $(PROVISION_DIR) $@

FORCE:

