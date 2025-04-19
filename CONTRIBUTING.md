# Contributing to Dead Man's Switch

Thank you for your interest in contributing to Dead Man's Switch! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Environment](#development-environment)
4. [Coding Standards](#coding-standards)
5. [Testing Guidelines](#testing-guidelines)
   - [Backend Testing](#backend-testing)
   - [Frontend Testing](#frontend-testing)
   - [Dynamic Coverage Requirements](#dynamic-coverage-requirements)
6. [Pull Request Process](#pull-request-process)
7. [Documentation](#documentation)

## Code of Conduct

Please be respectful and considerate of others when contributing to this project. We aim to foster an inclusive and welcoming community.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR-USERNAME/deadmanswitch.git`
3. Add the upstream repository: `git remote add upstream https://github.com/korjavin/deadmanswitch.git`
4. Create a new branch for your feature or bugfix: `git checkout -b feature/your-feature-name`

## Development Environment

We recommend using Docker/Podman for development to ensure consistency across environments:

```bash
# Using Docker
docker-compose up -d

# Using Podman
podman-compose up -d
```

The application will be available at http://localhost:8082.

## Coding Standards

- Follow Go's official [style guide](https://golang.org/doc/effective_go)
- Use meaningful variable and function names
- Write clear comments for complex logic
- Keep functions small and focused on a single responsibility
- Use proper error handling

## Testing Guidelines

We have a strong focus on testing to ensure the reliability and security of the application. All new features should include appropriate tests.

### Backend Testing

Run backend tests with:

```bash
go test ./...
```

For tests with coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### Mock Repositories

- Mock repositories should be placed in the `storage_test` package for reuse across test packages
- Use interfaces to allow for easy mocking of dependencies

### Frontend Testing

We use Playwright for end-to-end testing of the frontend.

#### Running Frontend Tests Locally

1. Make sure you have Node.js and npm installed
2. Install dependencies and Playwright browsers:

```bash
cd tests/frontend
npm install
npx playwright install
```

3. Run the tests:

```bash
# Using the provided script
./scripts/run-frontend-tests.sh

# Or manually
cd tests/frontend
npm test
```

4. View the test report:

```bash
npx playwright show-report
```

For more detailed information, see our [Frontend Testing Guide](docs/frontend-testing.md).

#### Running Frontend Tests in GitHub Actions

Frontend tests automatically run in GitHub Actions for all pull requests and pushes to the master branch. The workflow:

1. Builds the Docker image
2. Starts the application with a test database
3. Runs the Playwright tests against the running application
4. Uploads test reports and screenshots as artifacts

#### Writing a New Frontend Test Scenario

1. Create a new test file in `tests/frontend/` or add to an existing one
2. Follow this structure for your test:

```javascript
const { test, expect } = require('@playwright/test');

test.describe('Feature Name', () => {
  test('should perform expected action', async ({ page }) => {
    // 1. Setup - Register and/or login if needed
    await page.goto('http://localhost:8082/register');
    await page.fill('input[name="email"]', 'test@example.com');
    await page.fill('input[name="name"]', 'Test User');
    await page.fill('input[name="password"]', 'Password123!');
    await page.fill('input[name="confirmPassword"]', 'Password123!');
    await page.click('button[type="submit"]');

    // 2. Navigate to the feature you're testing
    await page.goto('http://localhost:8082/feature-page');

    // 3. Perform actions
    await page.click('#feature-button');
    await page.fill('input[name="feature-input"]', 'test value');
    await page.click('button[type="submit"]');

    // 4. Assert expected outcomes
    await expect(page.locator('.success-message')).toBeVisible();
    await expect(page.locator('.feature-result')).toContainText('Expected Result');

    // 5. Clean up (if necessary)
    await page.click('.logout-button');
  });
});
```

3. Best practices for frontend tests:
   - Start with a clean state (new user registration is preferred)
   - Test complete flows rather than isolated components
   - Use descriptive test names that explain what's being tested
   - Add comments to explain complex test steps
   - Only take screenshots on failure (configured in `playwright.config.js`)
   - Use page objects for complex pages to improve test maintainability

### Dynamic Coverage Requirements

We use a dynamic code coverage threshold that increases over time:

- Base threshold: 20% coverage
- Growth rate: +0.1% per commit
- Maximum threshold: 80% coverage

This approach allows us to gradually improve test coverage without blocking development. The formula is:

```
threshold = min(20 + (0.1 * commit_count), 80)
```

The CI pipeline will fail if the test coverage falls below this dynamic threshold.

For more details, see our [Dynamic Coverage Guide](docs/dynamic-coverage.md).

## Pull Request Process

1. Update the README.md and documentation with details of changes if appropriate
2. Ensure all tests pass locally before submitting
3. Update the CHANGELOG.md with details of changes
4. The PR will be merged once it receives approval from maintainers

## Documentation

- Update documentation for any new features or changes to existing functionality
- Document security considerations for security-related changes
- Keep the README.md up to date with new features and configuration options
