#!/bin/bash
# This script installs the necessary dependencies for the project.

set -e

# Color for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if go is installed
if ! command -v go &> /dev/null
then
    echo -e "${RED}Go could not be found. Please install Go and try again.${NC}"
    exit 1
fi

# Check if GOPATH is included in PATH
if [[ ":$PATH:" != *":$(go env GOPATH)/bin:"* ]]; then
    echo -e "${RED}GOPATH/bin is not in your PATH. Please add $(go env GOPATH)/bin to your PATH and try again.${NC}"
    exit 1
fi

# Install golangci-lint
if ! command -v golangci-lint &> /dev/null
then
    echo -e "${BLUE}Installing golangci-lint...${NC}"
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin
fi

# Install pre-commit
if ! command -v golangci-lint &> /dev/null
then
    echo -e "${BLUE}Installing pre-commit...${NC}"
    if command -v pipx &> /dev/null; then
        pipx install pre-commit
    else
        echo -e "${RED}pipx could not be found. Please install pipx and add this:\n\n  export PATH=\$PATH:/home/$(whoami)/.local/bin\n\nto your shell rc and try again.${NC}"
        exit 1
    fi
fi

# Install pre-commit hooks
echo -e "${BLUE}Installing pre-commit hooks...${NC}"
pre-commit install

echo -e "${GREEN}All dependencies installed successfully.${NC}"
