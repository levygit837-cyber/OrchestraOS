#!/usr/bin/env bash
set -euo pipefail

# install-tools.sh installs development tools required by the project.
# Run this once after cloning the repository.

GOLANGCI_VERSION="v1.64.6"

echo "Installing golangci-lint ${GOLANGCI_VERSION}..."
if command -v golangci-lint &> /dev/null; then
    echo "golangci-lint already installed: $(golangci-lint version)"
else
    # Binary install (requires curl)
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" "${GOLANGCI_VERSION}"
    echo "golangci-lint installed to $(go env GOPATH)/bin"
fi

echo "All tools installed."
