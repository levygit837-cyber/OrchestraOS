# OrchestraOS Makefile

.PHONY: all build test vet lint arch check install-tools format clean setup help

all: vet test arch lint build

## build: compile the entire project
build:
	go build ./...

## test: run all tests
test:
	go test ./... -race -count=1

## vet: run go vet on all packages
vet:
	go vet ./...

## lint: run golangci-lint (install with make install-tools if missing)
lint:
	./scripts/go/lint.sh

## arch: run architecture boundary tests
arch:
	go test ./tests/architecture/... -v -count=1

## check: run all checks (vet, test, arch, lint)
check: vet test arch lint build

## install-tools: install development tools (golangci-lint)
install-tools:
	./scripts/go/install-tools.sh

## format: format all Go files
format:
	go fmt ./...
	goimports -w .

## setup: install git hooks
setup:
	cp scripts/git/pre-commit.sh .git/hooks/pre-commit
	cp scripts/git/pre-push.sh .git/hooks/pre-push
	chmod +x .git/hooks/pre-commit .git/hooks/pre-push
	@echo "Git hooks installed."

## clean: remove build artifacts and temporary files
clean:
	rm -f orchestraos
	go clean -cache

## help: show this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //'
