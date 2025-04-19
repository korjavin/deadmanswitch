# Frontend Testing Guide

This document provides detailed information about writing and running frontend tests for the Dead Man's Switch application.

## Overview

We use [Playwright](https://playwright.dev/) for end-to-end testing of the frontend. Playwright allows us to automate browser interactions and verify that the application works correctly from a user's perspective.

## Test Structure

Our frontend tests follow a complete flow approach:

1. **Setup**: Register a new user and/or login
2. **Action**: Perform the actions being tested
3. **Verification**: Check that the expected results occurred
4. **Cleanup**: Logout or clean up any created resources

## Running Tests Locally

### Prerequisites

- Node.js (v14 or later)
- npm
- Docker or Podman with docker-compose/podman-compose

### Setup

1. Install dependencies:

```bash
cd tests/frontend
npm install
```

2. Install Playwright browsers:

```bash
npx playwright install
```

### Running Tests

You can run tests using the provided script:

```bash
./scripts/run-frontend-tests.sh
```

Or manually:

```bash
# Start the application in test mode
podman-compose -f docker-compose.yml -f docker-compose.override.yml up -d

# Run the tests
cd tests/frontend
npm test

# View the test report
npx playwright show-report
```

## Writing a New Test

### Test File Structure

Create a new test file in the `tests/frontend` directory with a `.spec.js` extension:

```javascript
// tests/frontend/feature-name.spec.js
const { test, expect } = require('@playwright/test');

test.describe('Feature Name', () => {
  test('should perform expected action', async ({ page }) => {
    // Test code here
  });
});
```

### Complete Example

Here's a complete example of a test that registers a user, creates a secret, and verifies it appears in the list:

```javascript
const { test, expect } = require('@playwright/test');

test.describe('Secret Management', () => {
  test('should create a new secret', async ({ page }) => {
    // 1. Register a new user
    await page.goto('http://localhost:8082/register');
    
    const email = `test-${Date.now()}@example.com`;
    await page.fill('input[name="email"]', email);
    await page.fill('input[name="name"]', 'Test User');
    await page.fill('input[name="password"]', 'Password123!');
    await page.fill('input[name="confirmPassword"]', 'Password123!');
    await page.click('button[type="submit"]');
    
    // Wait for registration to complete and redirect
    await expect(page).toHaveURL(/.*\/dashboard/);
    
    // 2. Navigate to secrets page
    await page.click('text=Secrets');
    await expect(page).toHaveURL(/.*\/secrets/);
    
    // 3. Create a new secret
    await page.click('text=Add Secret');
    
    const secretTitle = `Test Secret ${Date.now()}`;
    await page.fill('input[name="title"]', secretTitle);
    await page.fill('textarea[name="content"]', 'This is a test secret content');
    await page.click('button[type="submit"]');
    
    // 4. Verify the secret was created
    await expect(page.locator(`text=${secretTitle}`)).toBeVisible();
    
    // 5. Logout
    await page.click('text=Logout');
    await expect(page).toHaveURL(/.*\/login/);
  });
});
```

### Best Practices

1. **Use unique identifiers**: Generate unique emails and names using timestamps to avoid conflicts between test runs.

2. **Wait for navigation**: Always wait for page navigation to complete before proceeding.

3. **Use descriptive test names**: Make it clear what functionality is being tested.

4. **Test complete flows**: Test the entire user journey rather than isolated components.

5. **Clean up after tests**: Logout or clean up any resources created during the test.

6. **Use page objects for complex pages**: For complex pages, create page object classes to encapsulate page interactions:

```javascript
// page-objects/SecretPage.js
class SecretPage {
  constructor(page) {
    this.page = page;
  }

  async navigate() {
    await this.page.click('text=Secrets');
    await expect(this.page).toHaveURL(/.*\/secrets/);
  }

  async createSecret(title, content) {
    await this.page.click('text=Add Secret');
    await this.page.fill('input[name="title"]', title);
    await this.page.fill('textarea[name="content"]', content);
    await this.page.click('button[type="submit"]');
  }

  async verifySecretExists(title) {
    await expect(this.page.locator(`text=${title}`)).toBeVisible();
  }
}

module.exports = { SecretPage };
```

Then use it in your test:

```javascript
const { SecretPage } = require('./page-objects/SecretPage');

test('should create a new secret', async ({ page }) => {
  // Register and login...
  
  const secretsPage = new SecretPage(page);
  await secretsPage.navigate();
  
  const secretTitle = `Test Secret ${Date.now()}`;
  await secretsPage.createSecret(secretTitle, 'This is a test secret content');
  await secretsPage.verifySecretExists(secretTitle);
});
```

## Debugging Tests

To debug tests, run them with the `--debug` flag:

```bash
npm test -- --debug
```

This will run the tests in headed mode and pause execution, allowing you to see what's happening in the browser.

## CI/CD Integration

Our GitHub Actions workflow automatically runs frontend tests for all pull requests and pushes to the master branch. The workflow:

1. Builds the Docker image
2. Starts the application with a test database
3. Runs the Playwright tests against the running application
4. Uploads test reports and screenshots as artifacts

You can view test results in the GitHub Actions tab of the repository.

## Troubleshooting

### Common Issues

1. **Tests fail with "Navigation timeout"**: The application might be slow to start. Increase the timeout in the test or add explicit waits.

2. **Element not found**: The element might not be visible or might have a different selector. Use the `--debug` flag to see what's happening.

3. **Tests pass locally but fail in CI**: CI environments can be slower. Add more explicit waits or increase timeouts in the CI configuration.

### Getting Help

If you're having trouble with frontend tests, please:

1. Check the existing test files for examples
2. Review the [Playwright documentation](https://playwright.dev/docs/intro)
3. Open an issue with details about the problem you're encountering
