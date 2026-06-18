.PHONY: build go-test lint parity-audit test typecheck

##@ Quality
build: ## Build all TypeScript packages through Turbo
	$(call run_in,$(ROOT_PATH),pnpm build)

go-test: ## Run Go tests
	$(call run_in,$(ROOT_PATH)/packages/tempo-go,go test ./...)

lint: ## Run static checks
	$(call run_in,$(ROOT_PATH),pnpm lint)

parity-audit: ## Audit adjusted Tempo parity coverage
	$(call run_in,$(ROOT_PATH),pnpm --filter @alloy/infra parity:audit)

test: ## Run all TypeScript tests
	$(call run_in,$(ROOT_PATH),pnpm test)

typecheck: ## Typecheck all TypeScript packages
	$(call run_in,$(ROOT_PATH),pnpm typecheck)
