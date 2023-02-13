BDIR := $(shell cd -P "$$( dirname "$$SOURCE" )/../" && pwd )
CWD := $(shell cd -P "$$( dirname "$$SOURCE" )" && pwd )

####################################################################

default: menu
.PHONY: default

menu:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(lastword $(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: menu

stack: update ## initialize a new deployment
ifndef STACK
	$(error Use "STACK=mynewstack make stack" instead)
endif
	-pulumi stack init $(STACK)
.PHONY: stack

deploy: update ## deploy system stack
	#@TF_LOG=DEBUG pulumi --logtostderr -v=9 --config-file $(BDIR)/config/configuration.yml --non-interactive --cwd $(CWD) up -y 2> out.txt
	@pulumi --config-file $(BDIR)/config/configuration.yml --non-interactive --cwd $(CWD) up -y
.PHONY: deploy

destroy: update ## destroy the system stack
	@pulumi --config-file $(BDIR)/config/configuration.yml --non-interactive --cwd $(CWD) destroy -y
.PHONY: destroy

update:
	@go mod tidy
	@go mod download
.PHONY: update
