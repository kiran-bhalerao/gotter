
.PHONY: setup
## setup: add missing
setup:
		go mod tidy

.PHONY: test
## test: runs go test with default values,  or file=filename
test:
ifdef file

	 cd handlers && godotenv -f ../.env.test go test -goblin.run="$(file)" -goblin.timeout=300s
else

	 cd handlers && godotenv -f ../.env.test go test -goblin.timeout 300s
endif


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
