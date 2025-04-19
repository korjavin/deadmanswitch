const { test, expect } = require('@playwright/test');
const { TEST_USER, login, logout } = require('./utils');

test.describe('Authentication and User Menu Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Go to the home page before each test
    console.log('Navigating to home page...');
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    console.log('Home page loaded.');
  });

  // Take screenshot after each test
  test.afterEach(async ({ page }, testInfo) => {
    if (testInfo.status !== 'passed') {
      console.log(`Test failed: ${testInfo.title}`);
      await page.screenshot({ path: `test-failed-${testInfo.title.replace(/\s+/g, '-')}.png` });
    }
  });

  test('should show login and register links when not logged in', async ({ page }) => {
    // Check that login and register links are visible
    await expect(page.locator('a[href="/login"]')).toBeVisible();
    await expect(page.locator('a[href="/register"]')).toBeVisible();

    // Check that user dropdown is not visible
    await expect(page.locator('.dropdown-toggle')).not.toBeVisible();
  });

  test('should show user email in menu after login', async ({ page }) => {
    console.log('Starting login test...');

    // Take screenshot before login
    await page.screenshot({ path: 'before-login.png' });

    // Login with test credentials
    console.log('Logging in with test credentials...');
    await login(page, TEST_USER.email, TEST_USER.password);
    console.log('Login completed.');

    // Take screenshot after login
    await page.screenshot({ path: 'after-login.png' });

    // Check that the user email is displayed in the dropdown toggle
    console.log('Checking for user email in dropdown toggle...');
    const dropdownToggle = page.locator('.dropdown-toggle');
    await expect(dropdownToggle).toBeVisible();

    // Get the actual text for debugging
    const toggleText = await dropdownToggle.textContent();
    console.log(`Dropdown toggle text: "${toggleText}"`);

    // Check if it contains the email
    await expect(dropdownToggle).toContainText(TEST_USER.email);
    console.log('User email found in dropdown toggle.');

    // Check that login and register links are not visible
    console.log('Checking that login/register links are not visible...');
    await expect(page.locator('a[href="/login"]')).not.toBeVisible();
    await expect(page.locator('a[href="/register"]')).not.toBeVisible();

    // Cleanup - logout
    console.log('Logging out...');
    await logout(page);
    console.log('Logout completed.');
  });

  test('should show dropdown menu with profile and logout links when clicking on user email', async ({ page }) => {
    // Login with test credentials
    await login(page, TEST_USER.email, TEST_USER.password);

    // Click on the dropdown toggle (user email)
    await page.click('.dropdown-toggle');

    // Check that dropdown menu is visible
    const dropdownMenu = page.locator('.dropdown-menu');
    await expect(dropdownMenu).toBeVisible();

    // Check that profile and logout links are visible in the dropdown
    await expect(page.locator('a[href="/profile"]')).toBeVisible();
    await expect(page.locator('a[href="/settings"]')).toBeVisible();
    await expect(page.locator('a[href="/logout"]')).toBeVisible();

    // Cleanup - logout
    await logout(page);
  });

  test('should redirect to login page after logout', async ({ page }) => {
    // Login with test credentials
    await login(page, TEST_USER.email, TEST_USER.password);

    // Logout
    await logout(page);

    // Check that we're on the home page
    expect(page.url()).toContain('/');

    // Check that login and register links are visible again
    await expect(page.locator('a[href="/login"]')).toBeVisible();
    await expect(page.locator('a[href="/register"]')).toBeVisible();
  });
});
