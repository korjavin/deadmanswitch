/**
 * Setup file for Playwright tests
 * This file is executed once before all tests
 */

const { chromium } = require('@playwright/test');
const { TEST_USER } = require('./utils');

/**
 * Sleep for a specified number of milliseconds
 * @param {number} ms - Milliseconds to sleep
 */
const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

/**
 * Global setup function
 * Creates a test user if it doesn't exist
 */
async function globalSetup() {
  console.log('Starting test setup...');

  // Launch browser with slower navigation timeout
  const browser = await chromium.launch();
  const context = await browser.newContext({
    baseURL: 'http://localhost:8082',
    navigationTimeout: 60000, // 60 seconds timeout for navigation
  });
  const page = await context.newPage();

  try {
    console.log('Setting up test user...');

    // Wait for the server to be ready
    let serverReady = false;
    for (let i = 0; i < 10; i++) {
      try {
        console.log('Checking if server is ready...');
        await page.goto('/');
        serverReady = true;
        break;
      } catch (e) {
        console.log(`Server not ready yet, retrying in 3 seconds... (${i+1}/10)`);
        await sleep(3000);
      }
    }

    if (!serverReady) {
      throw new Error('Server did not become ready in time');
    }

    // First try to login with test credentials
    console.log('Attempting to login with test credentials...');
    await page.goto('/login');
    await page.fill('input[name="email"]', TEST_USER.email);
    await page.fill('input[name="password"]', TEST_USER.password);
    await page.click('button[type="submit"]');

    // Wait for navigation to complete
    await page.waitForLoadState('networkidle');

    // Check if login was successful by looking for dashboard URL or error message
    const url = page.url();
    if (url.includes('/dashboard')) {
      console.log('Test user already exists and login successful.');
    } else {
      console.log('Login failed. Attempting to register a new test user...');

      // Go to registration page
      await page.goto('/register');

      // Fill out registration form
      await page.fill('input[name="email"]', TEST_USER.email);
      await page.fill('input[name="name"]', TEST_USER.name);
      await page.fill('input[name="password"]', TEST_USER.password);
      await page.fill('input[name="confirmPassword"]', TEST_USER.password);

      // Take a screenshot before submitting
      await page.screenshot({ path: 'registration-form.png' });

      // Submit the form
      await page.click('button[type="submit"]');

      // Wait for navigation to complete
      await page.waitForLoadState('networkidle');

      // Take a screenshot after submitting
      await page.screenshot({ path: 'after-registration.png' });

      // Check if registration was successful
      const newUrl = page.url();
      if (newUrl.includes('/dashboard') || newUrl.includes('/login')) {
        console.log('Test user created successfully.');
      } else {
        console.log('Failed to create test user. Current URL:', newUrl);
        // Try to get any error messages
        const errorText = await page.textContent('body');
        console.log('Page content:', errorText.substring(0, 500) + '...');
      }
    }
  } catch (error) {
    console.error('Error during setup:', error);
    // Take a screenshot of the error state
    await page.screenshot({ path: 'setup-error.png' });
  } finally {
    // Close browser
    await browser.close();
    console.log('Setup completed.');
  }
}

module.exports = globalSetup;
