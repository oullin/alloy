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

