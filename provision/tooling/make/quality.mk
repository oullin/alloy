FMT_IMAGE := ghcr.io/oullin/go-fmt:v0.4.0
FMT_RUN := docker run --rm \
	--user $$(id -u):$$(id -g) \
	-v $(ROOT_PATH):/work \
	-w /work \
	-e HOST_PROJECT_PATH=$(ROOT_PATH) \
	$(FMT_IMAGE)

GO_GLOBS := '*.go'

define run_gofmt
@cd "$(ROOT_PATH)" && tmp=$$(mktemp); git ls-files -z $(1) --exclude-standard -- $(GO_GLOBS) | while IFS= read -r -d '' f; do [ -f "$$f" ] && printf '%s\0' "$$f"; done > "$$tmp"; if [ ! -s "$$tmp" ]; then echo "No Go files to format."; rm -f "$$tmp"; else xargs -0 $(FMT_RUN) go format < "$$tmp"; rc=$$?; rm -f "$$tmp"; exit $$rc; fi
endef

.PHONY: build format format-all go-test lint parity-audit test typecheck

##@ Quality
build: ## Build all TypeScript packages through Turbo
	$(call run_in,$(ROOT_PATH),pnpm build)

format: ## Format changed files with Prettier and gofmt
	$(call run_in,$(ROOT_PATH),pnpm exec prettier --write --cache .)
	$(call run_gofmt,--others --modified)

format-all: ## Format the full repository with Prettier and gofmt
	$(call run_in,$(ROOT_PATH),pnpm exec prettier --write .)
	$(call run_gofmt,--cached --others)

go-test: ## Run Go tests
	$(call run_in,$(ROOT_PATH)/packages/tempo/go,go test ./...)

lint: ## Run static checks
	$(call run_in,$(ROOT_PATH),pnpm lint)

parity-audit: ## Audit adjusted Tempo parity coverage
	$(call run_in,$(ROOT_PATH)/$(PROVISION_DIR),pnpm parity:audit)

test: ## Run all TypeScript tests
	$(call run_in,$(ROOT_PATH),pnpm test)

typecheck: ## Typecheck all TypeScript packages
	$(call run_in,$(ROOT_PATH),pnpm typecheck)
