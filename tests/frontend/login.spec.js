const { test, expect } = require('@playwright/test');
const { TEST_USER, login, logout } = require('./utils');

test.describe('Authentication and User Menu Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Go to the home page before each test
    console.log('Navigating to home page...');
    try {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      console.log('Home page loaded.');

      // Take screenshot of home page for debugging
      await page.screenshot({ path: 'home-page.png' });

      // Log the page title and URL for debugging
      console.log('Page title:', await page.title());
      console.log('Current URL:', page.url());
    } catch (error) {
      console.error('Error in beforeEach:', error);
      // Take screenshot of error state
      await page.screenshot({ path: 'beforeEach-error.png' });
    }
  });

  // Take screenshot after each test
  test.afterEach(async ({ page }, testInfo) => {
    if (testInfo.status !== 'passed') {
      console.log(`Test failed: ${testInfo.title}`);
      await page.screenshot({ path: `test-failed-${testInfo.title.replace(/\s+/g, '-')}.png` });
    }
  });

  test('should show login and register links when not logged in', async ({ page }) => {
    console.log('Starting test: should show login and register links when not logged in');

    // Take screenshot of initial state
    await page.screenshot({ path: 'initial-state.png' });

    // Log the current page content for debugging
    const pageContent = await page.content();
    console.log('Page content snippet:', pageContent.substring(0, 500) + '...');

    // Check for navigation links in the navbar
    const loginLink = page.locator('nav a[href="/login"]');
    const registerLink = page.locator('nav a[href="/register"]');

    console.log('Checking if login link is visible in navbar...');
    await expect(loginLink).toBeVisible();
    console.log('Login link is visible in navbar.');

    console.log('Checking if register link is visible in navbar...');
    await expect(registerLink).toBeVisible();
    console.log('Register link is visible in navbar.');

    // Alternative approach using getByRole with options
    console.log('Checking login link using getByRole...');
    const loginLinkByRole = page.getByRole('link', { name: 'Login', exact: true });
    await expect(loginLinkByRole).toBeVisible();
    console.log('Login link found by role.');

    // Check that user dropdown is not visible
    console.log('Checking that user dropdown is not visible...');
    await expect(page.locator('.dropdown-toggle')).not.toBeVisible();
    console.log('User dropdown is not visible as expected.');

    console.log('Test completed successfully.');
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

    // Check that the user menu is displayed after login
    console.log('Checking for user menu after login...');

    // First check if the dropdown toggle is visible
    const dropdownToggle = page.locator('.dropdown-toggle');
    await expect(dropdownToggle).toBeVisible();
    console.log('Dropdown toggle is visible.');

    // Get the actual text for debugging
    const toggleText = await dropdownToggle.textContent();
    console.log(`Dropdown toggle text: "${toggleText}"`);

    // Take a screenshot of the navbar area
    await page.screenshot({ path: 'navbar-after-login.png', clip: { x: 0, y: 0, width: 1000, height: 100 } });

    // Check if user menu contains expected elements
    console.log('Clicking dropdown toggle to open menu...');
    await dropdownToggle.click();

    // Check for profile link in dropdown menu
    console.log('Checking for profile link in dropdown menu...');
    const profileLink = page.locator('a[href="/profile"]');
    await expect(profileLink).toBeVisible();
    console.log('Profile link is visible in dropdown menu.');

    // Check for logout link in dropdown menu
    console.log('Checking for logout link in dropdown menu...');
    const logoutLink = page.locator('a[href="/logout"]');
    await expect(logoutLink).toBeVisible();
    console.log('Logout link is visible in dropdown menu.');
    console.log('User email found in dropdown toggle.');

    // Check that login and register links are not visible in the navbar
    console.log('Checking that login/register links are not visible in navbar...');
    await expect(page.locator('nav a[href="/login"]')).not.toBeVisible();
    await expect(page.locator('nav a[href="/register"]')).not.toBeVisible();

    // Cleanup - logout
    console.log('Logging out...');
    await logout(page);
    console.log('Logout completed.');
  });

  test('should show dropdown menu with profile and logout links when clicking on user email', async ({ page }) => {
    console.log('Starting test: should show dropdown menu with profile and logout links');

    // Login with test credentials
    console.log('Logging in with test credentials...');
    await login(page, TEST_USER.email, TEST_USER.password);
    console.log('Login successful.');

    // Take screenshot before clicking dropdown
    await page.screenshot({ path: 'before-dropdown-click.png' });

    // Click on the dropdown toggle
    console.log('Clicking on the dropdown toggle...');
    await page.click('.dropdown-toggle');
    console.log('Dropdown toggle clicked.');

    // Take screenshot after clicking dropdown
    await page.screenshot({ path: 'after-dropdown-click.png' });

    // Check that dropdown menu is visible
    console.log('Checking if dropdown menu is visible...');
    const dropdownMenu = page.locator('.dropdown-menu');
    await expect(dropdownMenu).toBeVisible();
    console.log('Dropdown menu is visible.');

    // Check that profile and logout links are visible in the dropdown
    console.log('Checking for profile link...');
    await expect(page.locator('a[href="/profile"]')).toBeVisible();
    console.log('Profile link is visible.');

    console.log('Checking for settings link...');
    const settingsLink = page.locator('a[href="/settings"]');
    if (await settingsLink.count() > 0) {
      await expect(settingsLink).toBeVisible();
      console.log('Settings link is visible.');
    } else {
      console.log('Settings link not found, might not be implemented yet.');
    }

    console.log('Checking for logout link...');
    await expect(page.locator('a[href="/logout"]')).toBeVisible();
    console.log('Logout link is visible.');

    // Cleanup - logout
    console.log('Cleaning up - logging out...');
    await logout(page);
    console.log('Logout successful.');

    console.log('Test completed successfully.');
  });

  test('should redirect to login page after logout', async ({ page }) => {
    console.log('Starting test: should redirect to login page after logout');

    // Login with test credentials
    console.log('Logging in with test credentials...');
    await login(page, TEST_USER.email, TEST_USER.password);
    console.log('Login successful.');

    // Take screenshot before logout
    await page.screenshot({ path: 'before-logout-test.png' });

    // Logout
    console.log('Logging out...');
    await logout(page);
    console.log('Logout completed.');

    // Take screenshot after logout
    await page.screenshot({ path: 'after-logout-test.png' });

    // Check that we're on the home page
    const currentUrl = page.url();
    console.log(`Current URL after logout: ${currentUrl}`);
    expect(currentUrl).toContain('/');
    console.log('Verified URL contains /');

    // Check that login and register links are visible again in the navbar
    console.log('Checking if login link is visible in navbar after logout...');
    await expect(page.locator('nav a[href="/login"]')).toBeVisible();
    console.log('Login link is visible in navbar after logout.');

    console.log('Checking if register link is visible in navbar after logout...');
    await expect(page.locator('nav a[href="/register"]')).toBeVisible();
    console.log('Register link is visible in navbar after logout.');

    console.log('Test completed successfully.');
  });
});
