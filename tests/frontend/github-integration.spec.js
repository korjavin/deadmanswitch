const { test, expect } = require('@playwright/test');

test.describe('GitHub Integration End-to-End Test', () => {
  // Increase test timeout to 2 minutes
  test.setTimeout(120000);
  // Generate a unique email for this test run
  const testEmail = `test-${Date.now()}@example.com`;
  const testPassword = 'Password123!';
  const testName = 'Test User';
  const githubUsername = 'korjavin';

  test('should register, connect and disconnect GitHub account', async ({ page }) => {
    console.log('Starting GitHub integration end-to-end test...');

    // Step 1: Register a new user
    console.log('Registering a new user...');
    await page.goto('http://localhost:8082/register');
    await page.fill('input[name="email"]', testEmail);
    await page.fill('input[name="name"]', testName);
    await page.fill('input[name="password"]', testPassword);
    await page.fill('input[name="confirmPassword"]', testPassword);

    await page.click('button[type="submit"]');

    // Wait for navigation to complete
    await page.waitForLoadState('networkidle', { timeout: 60000 });
    console.log('Registration form submitted');

    // Check if we're on the dashboard or login page
    const currentUrl = page.url();
    console.log(`Current URL after registration: ${currentUrl}`);

    // If we're on the login page, log in
    if (currentUrl.includes('/login')) {
      console.log('Redirected to login page, logging in...');
      await page.fill('input[name="email"]', testEmail);
      await page.fill('input[name="password"]', testPassword);
      await page.click('button[type="submit"]');
      await page.waitForLoadState('networkidle', { timeout: 60000 });
      console.log('Login form submitted');
    }

    // Check if we're on the dashboard
    const isDashboard = page.url().includes('/dashboard');
    console.log(`On dashboard: ${isDashboard}`);

    if (!isDashboard) {
      console.log('Not on dashboard, navigating to dashboard...');
      await page.goto('http://localhost:8082/dashboard');
      await page.waitForLoadState('networkidle', { timeout: 60000 });
    }

    // Step 2: Navigate to profile page
    console.log('Navigating to profile page...');
    await page.goto('http://localhost:8082/profile');
    await page.waitForLoadState('networkidle');
    console.log('Profile page loaded');

    // Step 3: Connect GitHub account
    console.log('Connecting GitHub account...');

    // Check if the GitHub section is visible
    const githubSection = page.locator('.card:has-text("GitHub Integration")');
    await expect(githubSection).toBeVisible({ timeout: 10000 });
    console.log('GitHub section is visible');

    // Check if we need to disconnect first
    const disconnectButton = page.locator('form[action="/profile/github/disconnect"] button');
    if (await disconnectButton.isVisible()) {
      console.log('GitHub already connected, disconnecting first...');
      await disconnectButton.click();
      await page.waitForLoadState('networkidle');
      console.log('GitHub disconnected');
    }

    // Now the form should be visible
    const githubUsernameInput = page.locator('#github_username');
    await expect(githubUsernameInput).toBeVisible({ timeout: 10000 });
    console.log('GitHub username input is visible');

    // Fill in GitHub username and submit
    await githubUsernameInput.fill(githubUsername);

    // Click the Connect GitHub button
    await page.click('form[action="/profile"] button:has-text("Connect GitHub")');

    // Wait for the form submission to complete
    await page.waitForLoadState('networkidle');
    console.log('GitHub connection form submitted');

    // Step 4: Verify GitHub connection
    console.log('Verifying GitHub connection...');

    // Reload the profile page
    await page.goto('http://localhost:8082/profile');
    await page.waitForLoadState('networkidle');

    // Check if the GitHub username is displayed
    const successAlert = page.locator('.alert-success:has-text("Your account is connected to GitHub")');
    await expect(successAlert).toBeVisible({ timeout: 10000 });

    const usernameText = page.locator('p:has-text("GitHub Username:")');
    await expect(usernameText).toContainText(githubUsername, { timeout: 10000 });
    console.log('GitHub connection verified');

    // Step 5: Disconnect GitHub account
    console.log('Disconnecting GitHub account...');

    // Check if the disconnect button is visible
    const disconnectBtn = page.locator('form[action="/profile/github/disconnect"] button');
    await expect(disconnectBtn).toBeVisible({ timeout: 10000 });

    // Click the disconnect button
    await disconnectBtn.click();
    await page.waitForLoadState('networkidle');
    console.log('GitHub disconnected');

    // Verify GitHub is disconnected
    await page.goto('http://localhost:8082/profile');
    await page.waitForLoadState('networkidle');

    // Check if the GitHub connection form is visible again
    const infoAlert = page.locator('.alert-info:has-text("Connect your GitHub account")');
    await expect(infoAlert).toBeVisible({ timeout: 10000 });

    const githubFormAfterDisconnect = page.locator('#github_username');
    await expect(githubFormAfterDisconnect).toBeVisible({ timeout: 10000 });
    console.log('GitHub disconnection verified');

    // Step 6: Logout
    console.log('Logging out...');
    await page.click('.dropdown-toggle');
    await page.click('a[href="/logout"]');
    await page.waitForLoadState('networkidle');
    console.log('Logout successful');

    console.log('GitHub integration end-to-end test completed');
  });
});
