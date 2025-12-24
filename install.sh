#!/bin/bash

# fogger installation script
# This script installs fogger on Linux and macOS systems

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Installing fogger - Cybersecurity tool for detecting gambling sites behind CDNs${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed. Please install Go 1.21 or higher first.${NC}"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | grep -o 'go[0-9]\.[0-9]*' | sed 's/go//')
MIN_VERSION="1.21"

if [ "$(printf '%s\n' "$MIN_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$MIN_VERSION" ]; then
    echo -e "${RED}Error: Go version $MIN_VERSION or higher is required. Current version: $GO_VERSION${NC}"
    exit 1
fi

echo -e "${GREEN}Go version $GO_VERSION detected${NC}"

# Check if Git is installed
if ! command -v git &> /dev/null; then
    echo -e "${RED}Error: Git is not installed.${NC}"
    exit 1
fi

echo -e "${GREEN}Git is available${NC}"

# Install fogger
echo -e "${YELLOW}Installing fogger...${NC}"

if command -v go &> /dev/null; then
    go install github.com/genesis410/fogger@latest
    echo -e "${GREEN}fogger installed successfully!${NC}"
else
    echo -e "${RED}Failed to install fogger${NC}"
    exit 1
fi

# Verify installation
if command -v fogger &> /dev/null; then
    echo -e "${GREEN}fogger is now available in your PATH${NC}"
    echo -e "${GREEN}Version information:${NC}"
    fogger --version 2>/dev/null || echo "Version flag not yet implemented"
else
    echo -e "${YELLOW}fogger was installed but is not in your PATH${NC}"
    echo -e "${YELLOW}You may need to add your Go bin directory to your PATH${NC}"
    echo -e "${YELLOW}Typically: $HOME/go/bin or GOPATH/bin${NC}"
fi

echo -e "${GREEN}Installation complete!${NC}"
echo -e "${GREEN}Get started with: fogger --help${NC}"