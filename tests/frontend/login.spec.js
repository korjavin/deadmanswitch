const { test, expect } = require('@playwright/test');
const { login, registerUser, logout } = require('./utils');

// Test data
const TEST_EMAIL = 'test@example.com';
const TEST_PASSWORD = 'Password123!';

test.describe('Authentication and User Menu Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Go to the home page before each test
    await page.goto('/');
  });

  test('should show login and register links when not logged in', async ({ page }) => {
    // Check that login and register links are visible
    await expect(page.locator('a[href="/login"]')).toBeVisible();
    await expect(page.locator('a[href="/register"]')).toBeVisible();
    
    // Check that user dropdown is not visible
    await expect(page.locator('.dropdown-toggle')).not.toBeVisible();
  });

  test('should show user email in menu after login', async ({ page }) => {
    // Login with test credentials
    await login(page, TEST_EMAIL, TEST_PASSWORD);
    
    // Check that the user email is displayed in the dropdown toggle
    const dropdownToggle = page.locator('.dropdown-toggle');
    await expect(dropdownToggle).toBeVisible();
    await expect(dropdownToggle).toContainText(TEST_EMAIL);
    
    // Check that login and register links are not visible
    await expect(page.locator('a[href="/login"]')).not.toBeVisible();
    await expect(page.locator('a[href="/register"]')).not.toBeVisible();
    
    // Cleanup - logout
    await logout(page);
  });

  test('should show dropdown menu with profile and logout links when clicking on user email', async ({ page }) => {
    // Login with test credentials
    await login(page, TEST_EMAIL, TEST_PASSWORD);
    
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
    await login(page, TEST_EMAIL, TEST_PASSWORD);
    
    // Logout
    await logout(page);
    
    // Check that we're on the home page
    expect(page.url()).toContain('/');
    
    // Check that login and register links are visible again
    await expect(page.locator('a[href="/login"]')).toBeVisible();
    await expect(page.locator('a[href="/register"]')).toBeVisible();
  });
});
