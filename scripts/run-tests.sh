#!/bin/bash

# Run tests with coverage for all packages
go test -coverprofile=coverage.out -covermode=atomic ./...

# Display coverage statistics
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Open the coverage report in the default browser (works on macOS, Linux with xdg-open, or Windows)
case "$(uname -s)" in
    Darwin)
        open coverage.html
        ;;
    Linux)
        if command -v xdg-open > /dev/null; then
            xdg-open coverage.html
        fi
        ;;
    CYGWIN*|MINGW*|MSYS*)
        start coverage.html
        ;;
    *)
        echo "Coverage report generated at coverage.html"
        ;;
esac

# Check if coverage is below threshold
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Total coverage: $COVERAGE%"
if (( $(echo "$COVERAGE < 20" | bc -l) )); then
    echo "Warning: Code coverage is below 20%"
    exit 1
fi
