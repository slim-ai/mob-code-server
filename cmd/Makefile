BDIR := $(shell cd -P "$$( dirname "$$SOURCE" )/../" && pwd )
CWD := $(shell cd -P "$$( dirname "$$SOURCE" )" && pwd )

####################################################################

default: menu
.PHONY: default

menu:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(lastword $(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: menu

stack: ## initialize a new deployment
	-pulumi stack init nomad
.PHONY: stack

deploy: ## deploy system stack
	@pulumi --config-file $(BDIR)/config/configuration.yml --non-interactive --cwd $(CWD) up -y
.PHONY: deploy

destroy: ## destroy the system stack
	@pulumi --config-file $(BDIR)/config/configuration.yml --non-interactive --cwd $(CWD) destroy -y
.PHONY: destroy
