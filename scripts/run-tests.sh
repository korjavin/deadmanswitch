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

# Calculate dynamic coverage threshold
# Get the number of commits
# First check if we have a shallow clone
if [ -f ".git/shallow" ]; then
    echo "Detected shallow clone, fetching complete history..."
    git fetch --unshallow || true
fi

COMMIT_COUNT=$(git rev-list --count HEAD)

# If we still have a shallow clone or the count is 1, use a fallback
if [ "$COMMIT_COUNT" = "1" ]; then
    # Fallback to a reasonable default
    COMMIT_COUNT=50
    echo "Could not determine accurate commit count, using default: $COMMIT_COUNT"
else
    echo "Total commits: $COMMIT_COUNT"
fi

# Calculate the dynamic threshold: 20% + 0.1% per commit, capped at 80%
THRESHOLD=$(echo "20 + 0.1 * $COMMIT_COUNT" | bc)
if (( $(echo "$THRESHOLD > 80" | bc -l) )); then
    THRESHOLD=80
fi
echo "Dynamic coverage threshold: $THRESHOLD%"

# Get the current coverage
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Total coverage: $COVERAGE%"

# Check if coverage is below the threshold
if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
    echo "Warning: Code coverage is below the dynamic threshold of $THRESHOLD%"
    exit 1
fi
