.PHONY: format format-all

##@ Formatting
format: ## Format + lint untracked + modified files via go-fmt v0.4.0 (oxfmt/oxlint)
	$(call run_fmt,--others --modified)

format-all: ## Format + lint every non-ignored file in the repo via go-fmt v0.4.0 (oxfmt/oxlint)
	@cd "$(ROOT_PATH)" && tmp=$$(mktemp); git ls-files -z --cached --others --exclude-standard -- $(GLOBS) | while IFS= read -r -d '' f; do [ -f "$$f" ] && printf '%s\0' "$$f"; done > "$$tmp"; if [ ! -s "$$tmp" ]; then echo "No files to format."; rm -f "$$tmp"; else xargs -0 $(FMT_RUN) format < "$$tmp"; rc=$$?; rm -f "$$tmp"; exit $$rc; fi
