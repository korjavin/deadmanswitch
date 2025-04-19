# Frontend Tests for Dead Man's Switch

This directory contains frontend tests for the Dead Man's Switch application using Playwright.

## Test Structure

- `setup.js`: Global setup file that creates a test user if it doesn't exist
- `utils.js`: Utility functions for common test operations (login, logout, etc.)
- `login.spec.js`: Tests for login functionality and user email display in menu

## Running Tests

To run the tests locally:

1. Make sure the application is running (using podman-compose):
   ```bash
   podman-compose up -d
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Install Playwright browsers:
   ```bash
   npx playwright install
   ```

4. Run the tests:
   ```bash
   npm test
   ```

5. Run the tests with browser UI visible:
   ```bash
   npm run test:headed
   ```

## Test Coverage

These tests cover:

- Checking if user email is displayed in the menu when logged in
- Verifying dropdown menu functionality
- Testing login and logout flows

## Adding New Tests

To add new tests:

1. Create a new test file in this directory with the `.spec.js` extension
2. Import the required utilities from `utils.js`
3. Write your tests using the Playwright API

## CI/CD Integration

These tests are automatically run as part of the GitHub Actions workflow in `.github/workflows/tests.yml`.
