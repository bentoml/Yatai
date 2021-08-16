.DEFAULT_GOAL := help

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


