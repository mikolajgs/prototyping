.DEFAULT_GOAL := help

.PHONY: help test

test: ## Runs tests
	cd pkg/rest-api && go test
	cd pkg/umbrella && go test
	cd pkg/ui && go test

run-example-app: ## Runs sample app
	docker rm -f sample-app-db
	docker run --name sample-app-db -d -e POSTGRES_PASSWORD=protopass -e POSTGRES_USER=protouser -e POSTGRES_DB=protodb -p 54320:5432 postgres:13
	sleep 10
	cd examples/sample_app && go build .
	cd examples/sample_app && ./sample_app

help: ## Displays this help
	@awk 'BEGIN {FS = ":.*##"; printf "$(MAKEFILE_NAME)\n\nUsage:\n  make \033[1;36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[1;36m%-25s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
