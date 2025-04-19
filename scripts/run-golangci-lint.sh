#!/bin/bash
set -e

# Function to check if golangci-lint is installed
check_golangci_lint() {
  GOLANGCI_LINT_PATH=$(which golangci-lint)
  
  if [ ! -f "$GOLANGCI_LINT_PATH" ]; then
    echo "golangci-lint is not installed at $GOLANGCI_LINT_PATH. Installing..."    
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
      # macOS - use the official installation script
      curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.2
    else
      # Linux and others
      curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.2
    fi
  fi
  
}

# Check if golangci-lint is installed, install if not
check_golangci_lint

# Print golangci-lint version
echo "Running golangci-lint $(golangci-lint --version | head -n 1)"

# Run golangci-lint with configuration file
golangci-lint run -c .golangci.yml