FMT_IMAGE := ghcr.io/oullin/go-fmt:v0.4.0
FMT_RUN := docker run --rm \
	--user $$(id -u):$$(id -g) \
	-v $(ROOT_PATH):/work \
	-w /work \
	-e HOST_PROJECT_PATH=$(ROOT_PATH) \
	$(FMT_IMAGE)

GLOBS := '*.go' '*.ts' '*.tsx' '*.vue' '*.mts' '*.cts'

define run_fmt
@cd "$(ROOT_PATH)" && tmp=$$(mktemp); git ls-files -z $(1) --exclude-standard -- $(GLOBS) | while IFS= read -r -d '' f; do [ -f "$$f" ] && printf '%s\0' "$$f"; done > "$$tmp"; if [ ! -s "$$tmp" ]; then echo "No files to format."; rm -f "$$tmp"; else xargs -0 $(FMT_RUN) format < "$$tmp"; rc=$$?; rm -f "$$tmp"; exit $$rc; fi
endef

define step
	@printf "\n\033[1;36m==>\033[0m \033[1m%s\033[0m\n" "$(1)"
endef

define run_in
	@cd "$(1)" && $(2)
endef

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*?## "; printf "\nUsage: \033[36mmake <target>\033[0m\n"} \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5); next } \
		/^[a-zA-Z0-9_.-]+:.*?##/ { printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

