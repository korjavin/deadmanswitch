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
MAX_WAIT=120
COUNTER=0

# Print container status
echo "Container status:"
podman ps

# Print network status
echo "Network status:"
podman network ls

# Wait for the application to start
while [ $COUNTER -lt $MAX_WAIT ]; do
  # Check if container is running
  if ! podman ps | grep -q deadmanswitch_deadmanswitch; then
    echo "Container is not running! Checking logs..."
    podman logs deadmanswitch_deadmanswitch_1 || true
    echo "Trying to restart container..."
    podman-compose up -d
    sleep 5
  fi

  # Try to connect to the application
  STATUS=$(check_app_ready)
  if [ "$STATUS" = "200" ]; then
    echo "Application is ready!"
    break
  fi

  COUNTER=$((COUNTER+1))
  echo "Waiting for application to start... ($((MAX_WAIT-COUNTER)) seconds left)"

  # Print more detailed status every 10 seconds
  if [ $((COUNTER % 10)) -eq 0 ]; then
    echo "Current container status:"
    podman ps
    echo "Checking container logs (last 10 lines):"
    podman logs --tail=10 deadmanswitch_deadmanswitch_1 || true
  fi

  sleep 1
done

# Check if we timed out
if [ $COUNTER -eq $MAX_WAIT ]; then
  echo "Timed out waiting for application to start"
  # Print logs to help debug
  echo "Container logs:"
  podman logs deadmanswitch_deadmanswitch_1 || true
  echo "Container status:"
  podman ps
  echo "Network status:"
  podman network ls
  exit 1
fi

# Run the tests using the Playwright Docker image
echo "Running tests with Playwright Docker image..."
cd "$PROJECT_ROOT"

# Check if we're running in CI
if [ -n "$CI" ]; then
  echo "Running in CI environment"
  # In CI, use the Docker image
  podman-compose -f "$PROJECT_ROOT/tests/frontend/docker-compose.playwright.yml" run --rm playwright
  TEST_EXIT_CODE=$?
else
  echo "Running in local environment"
  # Locally, run directly
  cd "$PROJECT_ROOT/tests/frontend"
  npx playwright test "$@"
  TEST_EXIT_CODE=$?
fi

# Stop the containers
echo "Stopping containers..."
cd "$PROJECT_ROOT"
podman-compose down

# Exit with the test exit code
exit $TEST_EXIT_CODE
