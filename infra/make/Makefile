SHELL := /bin/bash

INFRA_MAKE_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
ROOT_PATH := $(abspath $(INFRA_MAKE_DIR)/../..)
INFRA_MAKE_DIR := $(ROOT_PATH)/infra/make

# Shared helpers/vars first (defines run_in, paths, etc.), then each concern.
# Build and quality orchestration lives under infra/make so target definitions
# have one owner.
include $(INFRA_MAKE_DIR)/shared.mk
include $(INFRA_MAKE_DIR)/quality.mk
include $(INFRA_MAKE_DIR)/format.mk

.DEFAULT_GOAL := help
