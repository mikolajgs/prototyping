.DEFAULT_GOAL := help

.PHONY: help test

test: ## Runs tests
	cd pkg/struct-sql-postgres && go test
	cd pkg/struct2db && go test
	cd pkg/restapi && go test
	cd pkg/ui && go test

run-sample-app: ## Runs sample app
	docker rm -f sample-app-db
	docker run --name sample-app-db -d -e POSTGRES_PASSWORD=uipass -e POSTGRES_USER=uiuser -e POSTGRES_DB=uidb -p 54320:5432 postgres:13
	sleep 5
	cd cmd/sample_app && go build .
	cd cmd/sample_app && ./sample_app

help: ## Displays this help
	@awk 'BEGIN {FS = ":.*##"; printf "$(MAKEFILE_NAME)\n\nUsage:\n  make \033[1;36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[1;36m%-25s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
