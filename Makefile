default: menu
.PHONY: default

menu:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(lastword $(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: menu

#######################################################################
# PRIMARY TARGETS

stack: vendor ## provision a new code server stack
	@$(MAKE) -C cmd stack
.PHONY: stack

deploy: vendor ## deploy the new code server stack
	@$(MAKE) -C cmd deploy
.PHONY: deploy

destroy: ## deprovision the code server stack
	@$(MAKE) -C cmd destroy
.PHONY: destroy

tools: ## installs or upgrades needed tools
	@bash scripts/tools.sh
.PHONY: tools

vendor:
	@go mod vendor
.PHONY: vendor
