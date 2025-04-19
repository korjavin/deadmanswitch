#!/bin/bash

# Make script exit on first error
set -e

# Check if podman-compose is installed
if ! command -v podman-compose &> /dev/null; then
    echo "Error: podman-compose is not installed. Please install it first."
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "Error: npm is not installed. Please install Node.js and npm first."
    exit 1
fi

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    npm install
fi

# Install Playwright browsers if needed
if [ ! -d "$HOME/.cache/ms-playwright" ]; then
    echo "Installing Playwright browsers..."
    npx playwright install
fi

# Start the application with the test configuration
echo "Starting the application with test configuration..."
podman-compose -f docker-compose.yml -f docker-compose.override.yml up -d

# Wait for the application to start
echo "Waiting for the application to start..."
sleep 5

# Run the tests
echo "Running frontend tests..."
npm test -- --debug

# Show the test report
echo "Opening test report..."
npx playwright show-report

# Ask if user wants to stop the application
read -p "Do you want to stop the application? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Stopping the application..."
    podman-compose -f docker-compose.yml -f docker-compose.override.yml down
fi

echo "Done!"
