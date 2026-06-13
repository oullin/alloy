.PHONY: oracle-install oracle-generate oracle-check

ORACLE_PATH := $(ROOT_PATH)/packages/carbon-oracle

##@ Oracle
oracle-install: ## Verify oracle workspace dependencies are installed
	$(call run_in,$(ROOT_PATH),pnpm install --frozen-lockfile)

oracle-generate: oracle-install ## Regenerate oracle fixtures
	$(call run_in,$(ORACLE_PATH),pnpm oracle:generate)

oracle-check: oracle-install ## Verify committed fixtures match the oracle generator
	$(call run_in,$(ORACLE_PATH),pnpm oracle:check)
