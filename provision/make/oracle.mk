.PHONY: oracle-install oracle-generate oracle-check

ORACLE_PATH := $(ROOT_PATH)/packages/carbon-oracle

##@ Oracle
oracle-install: ## Install PHP oracle dependencies
	$(call run_in,$(ORACLE_PATH),composer install --no-interaction --prefer-dist)

oracle-generate: oracle-install ## Regenerate Carbon oracle fixtures
	$(call run_in,$(ORACLE_PATH),php src/generate.php)

oracle-check: oracle-install ## Verify committed fixtures match the Carbon oracle
	$(call run_in,$(ORACLE_PATH),php src/generate.php --check)

