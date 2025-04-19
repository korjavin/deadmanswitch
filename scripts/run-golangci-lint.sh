#!/bin/bash
set -e

# Function to check if golangci-lint is installed
check_golangci_lint() {
  GOLANGCI_LINT_PATH=$(go env GOPATH)/bin/golangci-lint
  
  if [ ! -f "$GOLANGCI_LINT_PATH" ]; then
    echo "golangci-lint is not installed at $GOLANGCI_LINT_PATH. Installing..."
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
      # macOS - use the official installation script
      curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
    else
      # Linux and others
      curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
    fi
  fi
  
  # Export the GOPATH/bin to PATH to ensure we can use golangci-lint
  export PATH=$(go env GOPATH)/bin:$PATH
}

# Check if golangci-lint is installed, install if not
check_golangci_lint

# Print golangci-lint version
echo "Running golangci-lint $($(go env GOPATH)/bin/golangci-lint --version | head -n 1)"

# Run golangci-lint
$(go env GOPATH)/bin/golangci-lint run