SHELL := /bin/bash

INFRA_MAKEFILE := infra/make/Makefile

.DEFAULT_GOAL := help

# Every real target lives in infra/make. This root Makefile is a transparent
# forwarder so existing `make <target>` commands keep working.
.PHONY: help FORCE
help:
	@$(MAKE) --no-print-directory -f $(INFRA_MAKEFILE) help

Makefile:
	@:

%: FORCE
	@$(MAKE) --no-print-directory -f $(INFRA_MAKEFILE) $@

FORCE:
