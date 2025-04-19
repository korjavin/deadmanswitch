/**
 * Setup file for Playwright tests
 * This file is executed once before all tests
 */

const { chromium } = require('@playwright/test');
const { login, registerUser } = require('./utils');

/**
 * Global setup function
 * Creates a test user if it doesn't exist
 */
async function globalSetup() {
  // Test data
  const TEST_EMAIL = 'test@example.com';
  const TEST_PASSWORD = 'Password123!';

  // Launch browser
  const browser = await chromium.launch();
  const page = await browser.newPage();

  try {
    // Try to login with test credentials
    await page.goto('/login');
    await page.fill('input[name="email"]', TEST_EMAIL);
    await page.fill('input[name="password"]', TEST_PASSWORD);
    await page.click('button[type="submit"]');

    // Check if login was successful
    const url = page.url();
    if (!url.includes('/dashboard')) {
      // If login failed, register a new user
      console.log('Test user does not exist. Creating a new test user...');
      await registerUser(page, TEST_EMAIL, TEST_PASSWORD);
      console.log('Test user created successfully.');
    } else {
      console.log('Test user already exists.');
    }
  } catch (error) {
    console.error('Error during setup:', error);
  } finally {
    // Close browser
    await browser.close();
  }
}

module.exports = globalSetup;
