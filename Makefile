# OrchestraOS Makefile
# Standard targets for development, testing, and module creation.

.PHONY: all build test vet lint arch contracts check check-imports install-tools clean new-module help

all: vet test arch contracts lint build

## build: compile the entire project
build:
	go build ./...

## test: run all tests (unit, integration, architecture)
test:
	go test ./... -race -count=1

## vet: run go vet on all packages
vet:
	go vet ./...

## lint: run golangci-lint (install with make install-tools if missing)
lint:
	./scripts/lint.sh

## arch: run architecture boundary tests
arch:
	go test ./tests/architecture/... -v -count=1

## check-imports: verify no cross-module imports exist
check-imports:
	go test ./tests/architecture/... -run TestModuleBoundaries -v
	go test ./tests/architecture/... -run TestModulesDoNotImportOrchestration -v
	go test ./tests/architecture/... -run TestTransitionPackageIsLeaf -v
	go test ./tests/architecture/... -run TestOnlyOrchestrationImportsModules -v

## contracts: verify CONTRACTS.md sync with code
contracts:
	./scripts/verify-contracts.sh

## check: run all checks (vet, test, arch, contracts, lint)
check: vet test arch contracts lint size-check

## install-tools: install development tools (golangci-lint)
install-tools:
	./scripts/install-tools.sh

## new-module: create a new module from template (usage: make new-module NAME=foo)
new-module:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make new-module NAME=<module-name>"; \
		exit 1; \
	fi
	./scripts/new-module.sh "$(NAME)"

## size-check: check if any module exceeds the recommended size limit (informational)
size-check:
	-@./scripts/check_module_size.sh || true

## format: format all Go files
format:
	go fmt ./...
	goimports -w .

## clean: remove build artifacts and temporary files
clean:
	rm -f orchestraos
	go clean -cache

## help: show this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //'
