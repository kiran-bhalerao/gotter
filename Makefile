
.PHONY: setup
## setup: add missing
setup:
		go mod tidy

.PHONY: test
## test: runs go test with default values
test:
	 cd handlers && godotenv -f ../.env.test go test

.PHONY: run
## run: runs go run *.go
run:
	godotenv -f ./.env go run *.go

.PHONY: watch
## watch: runs go run *.go in watch mode
watch:
	nodemon --exec godotenv -f ./.env go run *.go --signal SIGTERM

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
