.DEFAULT_GOAL := help

help: ## Show all Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

be-deps: ## Fetch Golang deps
	echo "Downloading go modules..."
	go mod download

be-run: ## Start backend API server
	@echo "Make sure to install postgresql and create yatai DB with 'createdb yatai'"
	if [[ ! -f ./yatai-config.dev.yaml ]]; then \
		echo "yatai-config.dev.yaml not found. Creating one with postgresql user: " $$(whoami); \
		cp ./yatai-config.sample.yaml ./yatai-config.dev.yaml; \
		sed -i 's/user: .*/user: '$$(whoami)'/' ./yatai-config.dev.yaml; \
	fi; \
	go run ./api-server/main.go serve -d -c ./yatai-config.dev.yaml

fe-deps: ## Fetch react deps
	cd ui && yarn

fe-run: ## Run frontend components
	cd ui && yarn start


