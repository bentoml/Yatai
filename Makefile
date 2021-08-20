.DEFAULT_GOAL := help

VERSION := $(shell git describe --match=NeVeRmAtCh --tags --always --dirty | cut -c 1-7)
DOCKER_REGISTRY := 192023623294.dkr.ecr.ap-northeast-1.amazonaws.com

BUILDER_IMG := $(DOCKER_REGISTRY)/yatai-builder:1.0
UI_BUILDER_IMG := $(DOCKER_REGISTRY)/yatai-ui-builder:1.0
YATAI_IMG := $(DOCKER_REGISTRY)/yatai:$(VERSION)

GOMOD_CACHE ?= "$(GOPATH)/pkg/mod"

BASE_CNTR_ARGS := -u root \
	--privileged --rm --net=host \
	-e GOPROXY=https://goproxy.io \
	-v ${GOMOD_CACHE}:/go/pkg/mod \
	-v $(realpath /etc/localtime):/etc/localtime:ro \
	-v $(PWD):/code \
	-w /code

BUILDER_CNTR_ARGS := $(BASE_CNTR_ARGS) $(BUILDER_IMG)
BUILDER_CNTR_CMD := docker run $(BUILDER_CNTR_ARGS)
BUILDER_CNTR_TTY_CMD := docker run -it $(BUILDER_CNTR_ARGS)

UI_BUILDER_CNTR_ARGS := $(BASE_CNTR_ARGS) $(UI_BUILDER_IMG)
UI_BUILDER_CNTR_CMD := docker run $(UI_BUILDER_CNTR_ARGS)
UI_BUILDER_CNTR_TTY_CMD := docker run -it $(UI_BUILDER_CNTR_ARGS)

pull-ui-builder-image:
	docker pull $(UI_BUILDER_IMG) || true

pull-builder-image:
	docker pull $(BUILDER_IMG) || true

docker-build-ui: pull-ui-builder-image
	$(UI_BUILDER_CNTR_CMD) sh -c "cd dashboard; ln -s /cache/node_modules ./node_modules; yarn build"
	echo "build ui done"

docker-golint: pull-builder-image
	$(BUILDER_CNTR_CMD) ./scripts/ci/golint.sh

docker-gofmt-chk: pull-builder-image
	$(BUILDER_CNTR_CMD) ./scripts/ci/gofmt-check.sh

docker-gofmt-fmt: pull-builder-image
	$(BUILDER_CNTR_CMD) ./scripts/ci/gofmt.sh -w

docker-eslint: pull-ui-builder-image
	$(UI_BUILDER_CNTR_CMD) sh -c "cd dashboard; ln -s /cache/node_modules ./node_modules; yarn lint"

docker-ui-typecheck: pull-ui-builder-image
	$(UI_BUILDER_CNTR_CMD) sh -c "cd dashboard; ln -s /cache/node_modules ./node_modules; yarn typecheck"

docker-build-api-server: pull-builder-image
	$(BUILDER_CNTR_CMD) sh -c "mkdir -p ./bin; go build -o ./bin/api-server ./api-server/main.go"

build-builder-image:
	docker build -f Dockerfile-builder -t $(BUILDER_IMG) . || exit 1
	docker push $(BUILDER_IMG)

build-ui-builder-image: pull-ui-builder-image
	docker build -f Dockerfile-ui-builder -t $(UI_BUILDER_IMG) . || exit 1
	docker push $(UI_BUILDER_IMG)

build-image:
	docker build -t $(YATAI_IMG) .
	docker push $(YATAI_IMG)

build: docker-build-ui docker-build-api-server build-image

ui-builder-cli:
	$(UI_BUILDER_CNTR_TTY_CMD) sh

help: ## Show all Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

yatai-dev: ## Run yatai(be and fe) in development mode
	@make -j2 be-run fe-run
yatai-d: ## Build docker images
	@docker build -t yatai:production .
yatai-d-r: ## Run docker image
	@docker run -it -p 3000:3000 -p 7777:7777 yatai:production

be-deps: ## Fetch Golang deps
	@echo "Downloading go modules..."
	@go mod download
be-build: ## Build backend binary
	@go build -o ./bin/yatai-api-server ./api-server/main.go
be-run: be-build ## Start backend API server
	@echo "Make sure to install postgresql and create yatai DB with 'createdb yatai'"
	@if [[ ! -f ./yatai-config.dev.yaml ]]; then \
		echo "yatai-config.dev.yaml not found. Creating one with postgresql user: " $$(whoami); \
		cp ./yatai-config.sample.yaml ./yatai-config.dev.yaml; \
		sed -i 's/user: .*/user: '$$(whoami)'/' ./yatai-config.dev.yaml; \
	fi; \
	./bin/yatai-api-server serve -d -c ./yatai-config.dev.yaml

fe-deps: ## Fetch frontend deps
	@cd dashboard && yarn
fe-build: ## Build frontend for production
	@cd dashboard && yarn build
fe-run: ## Run frontend components
	@cd dashboard && yarn start


