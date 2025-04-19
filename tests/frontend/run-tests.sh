#!/bin/bash

# Get the absolute path to the project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

# Generate a unique test run ID
export TEST_RUN_ID=$(date +%s)
echo "Using test database with ID: $TEST_RUN_ID"

# Stop any running containers
echo "Stopping existing containers..."
podman-compose down

# Start the application with the test configuration
echo "Starting application with test configuration..."
podman-compose -f docker-compose.yml -f "$PROJECT_ROOT/tests/frontend/docker-compose.test.yml" up -d

# Wait for the application to start
echo "Waiting for application to start..."
sleep 5

# Run the tests
echo "Running tests..."
cd "$PROJECT_ROOT/tests/frontend"
npx playwright test "$@"
TEST_EXIT_CODE=$?

# Stop the containers
echo "Stopping containers..."
cd "$PROJECT_ROOT"
podman-compose down

# Exit with the test exit code
exit $TEST_EXIT_CODE
