.DEFAULT_GOAL := help

GIT_COMMIT := $(shell git describe --match=NeVeRmAtCh --tags --always --dirty | cut -c 1-7)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION := $(shell git describe --tags `git rev-list --tags --max-count=1` | sed 's/v\(\)/\1/')

PKG := github.com/bentoml/yatai
VERSION_BUILDFLAGS := -X '$(PKG)/api-server/version.GitCommit=$(GIT_COMMIT)' -X '$(PKG)/api-server/version.Version=$(VERSION)' -X '$(PKG)/api-server/version.BuildDate=$(BUILD_DATE)'
DOCKER_REGISTRY := quay.io/bentoml

BUILDER_IMG := $(DOCKER_REGISTRY)/yatai-builder:1.0
UI_BUILDER_IMG := $(DOCKER_REGISTRY)/yatai-ui-builder:1.0
YATAI_IMG := $(DOCKER_REGISTRY)/yatai:$(GIT_COMMIT)

GOMOD_CACHE ?= "$(GOPATH)/pkg/mod"

BASE_CNTR_ARGS := -u root \
	--privileged --rm --net=host \
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

docker-build-ui: pull-ui-builder-image ## Docker build UI
	$(UI_BUILDER_CNTR_CMD) sh -c "cd dashboard; ln -s /cache/node_modules ./node_modules; yarn build"
	echo "build ui done"

docker-golint: pull-builder-image ## Docker golint
	$(BUILDER_CNTR_CMD) ./scripts/ci/golint.sh

docker-gofmt-chk: pull-builder-image ## Docker gofmt-check
	$(BUILDER_CNTR_CMD) ./scripts/ci/gofmt-check.sh

docker-gofmt-fmt: pull-builder-image ## Docker gofmt
	$(BUILDER_CNTR_CMD) ./scripts/ci/gofmt.sh -w

docker-eslint: pull-ui-builder-image ## Docker eslint
	$(UI_BUILDER_CNTR_CMD) sh -c "cd dashboard; ln -s /cache/node_modules ./node_modules; yarn lint"

docker-ui-typecheck: pull-ui-builder-image ## Docker typecheck
	$(UI_BUILDER_CNTR_CMD) sh -c "cd dashboard; ln -s /cache/node_modules ./node_modules; yarn typecheck"

docker-build-api-server: pull-builder-image ## Build api-server binary
	mkdir -p ./bin
	$(BUILDER_CNTR_CMD) go build -ldflags "$(VERSION_BUILDFLAGS)" -o ./bin/api-server ./api-server/main.go

build-api-server:
	mkdir -p ./bin
	go build -ldflags "$(VERSION_BUILDFLAGS)" -o ./bin/api-server ./api-server/main.go

build-builder-image: ## Build builder image
	docker build -f Dockerfile-builder -t $(BUILDER_IMG) . || exit 1
	docker push $(BUILDER_IMG)

build-ui-builder-image: pull-ui-builder-image ## Build UI builder image
	docker build -f Dockerfile-ui-builder -t $(UI_BUILDER_IMG) . || exit 1
	docker push $(UI_BUILDER_IMG)

build-image: ## Build Yatai image
	docker build -t $(YATAI_IMG) .
	docker push $(YATAI_IMG)

tag-release: ## Tag Yatai image as release
	docker tag $(YATAI_IMG) $(DOCKER_REGISTRY)/yatai:$(VERSION)
	docker push $(DOCKER_REGISTRY)/yatai:$(VERSION)

build: docker-build-ui docker-build-api-server build-image ## Build pipeline

ui-builder-cli:
	$(UI_BUILDER_CNTR_TTY_CMD) sh

help: ## Show all Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

yatai-dev: ## Run yatai(be and fe) in development mode
	@make -j2 be-run fe-run

be-deps: ## Fetch Golang deps
	@echo "Downloading go modules..."
	@go mod download
be-run:
	@echo "Make sure to install postgresql and create yatai DB with 'createdb yatai'"
	@if [[ ! -f ./yatai-config.dev.yaml ]]; then \
		echo "yatai-config.dev.yaml not found. Creating one with postgresql user: " $$(whoami); \
		cp ./yatai-config.sample.yaml ./yatai-config.dev.yaml; \
		sed -i 's/user: .*/user: '$$(whoami)'/' ./yatai-config.dev.yaml; \
	fi; \
	go run -ldflags "$(VERSION_BUILDFLAGS)" ./api-server/main.go version
	go run -ldflags "$(VERSION_BUILDFLAGS)" ./api-server/main.go serve -d -c ./yatai-config.dev.yaml

fe-deps: ## Fetch frontend deps
	@cd dashboard && yarn
fe-build: ## Build frontend for production
	@cd dashboard && yarn build
fe-run: ## Run frontend components
	@cd dashboard && yarn start

