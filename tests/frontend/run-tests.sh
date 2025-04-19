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

# Function to check if the application is ready
check_app_ready() {
  curl -s -o /dev/null -w "%{http_code}" http://localhost:8082/login
}

# Wait for the application to be ready with a timeout
MAX_WAIT=60
COUNTER=0
while [ $COUNTER -lt $MAX_WAIT ]; do
  STATUS=$(check_app_ready)
  if [ "$STATUS" = "200" ]; then
    echo "Application is ready!"
    break
  fi
  COUNTER=$((COUNTER+1))
  echo "Waiting for application to start... ($((MAX_WAIT-COUNTER)) seconds left)"
  sleep 1
done

# Check if we timed out
if [ $COUNTER -eq $MAX_WAIT ]; then
  echo "Timed out waiting for application to start"
  # Print logs to help debug
  echo "Container logs:"
  podman logs deadmanswitch_deadmanswitch_1
  exit 1
fi

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
