.DEFAULT_GOAL := help

ifndef VERSION
	VERSION := "0.0.1-dummy.1"
endif

ifndef CLONE_DIR
	CLONE_DIR := $(shell mktemp -d)
endif

help: ## Show all Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

lint:
	ct lint --all

cleanup:
	rm yatai-*.tgz || true

release: cleanup
	helm package . --version ${VERSION} --app-version ${VERSION}
	if [ ${GITHUB_ACTIONS} ]; then git config --global user.name "github-actions[bot]"; fi
	if [ ${GITHUB_ACTIONS} ]; then git config --global user.email "github-actions[bot]@users.noreply.github.com"; fi
	if [ ${GITHUB_ACTIONS} ]; then git config --global credential.helper store; fi
	git clone --single-branch --branch gh-pages "https://x-access-token:${API_TOKEN_GITHUB}@github.com/bentoml/charts.git" "${CLONE_DIR}"
	cp -R yatai-${VERSION}.tgz ${CLONE_DIR}
	cd ${CLONE_DIR}; helm repo index .
	cd ${CLONE_DIR}; git add . && git commit --message "release yatai ${VERSION}" && git push -f origin HEAD:gh-pages || exit 1
	rm -rf ${CLONE_DIR} || true

template:
	helm template yatai ./ -n yatai-system --debug

