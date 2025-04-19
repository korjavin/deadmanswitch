/**
 * Utility functions for frontend tests
 */

// Test data that can be reused across tests
const TEST_USER = {
  email: 'korjavin@gmail.com',
  password: 'password',
  name: 'Test User'
};

/**
 * Login to the application
 * @param {import('@playwright/test').Page} page - Playwright page
 * @param {string} email - User email
 * @param {string} password - User password
 */
async function login(page, email, password) {
  console.log(`Navigating to login page...`);
  await page.goto('http://localhost:8082/login');
  await page.waitForLoadState('networkidle');

  console.log(`Filling login form with email: ${email}...`);
  await page.fill('input[name="email"]', email);
  await page.fill('input[name="password"]', password);

  // Take screenshot before submitting
  await page.screenshot({ path: 'login-form-filled.png' });

  console.log('Submitting login form...');
  await page.click('button[type="submit"]');

  // Wait for navigation to complete
  console.log('Waiting for navigation to complete...');
  await page.waitForLoadState('networkidle');

  // Take screenshot after login attempt
  await page.screenshot({ path: 'after-login-attempt.png' });

  // Verify we're on the dashboard
  const currentUrl = page.url();
  console.log(`Current URL after login: ${currentUrl}`);

  if (!currentUrl.includes('/dashboard')) {
    console.log('Login failed. Checking for error messages...');
    const errorText = await page.textContent('body');
    console.log('Page content:', errorText.substring(0, 500) + '...');
    throw new Error(`Login failed. Current URL: ${currentUrl}`);
  }

  console.log('Login successful.');
}

/**
 * Register a new user
 * @param {import('@playwright/test').Page} page - Playwright page
 * @param {string} email - User email
 * @param {string} name - User name
 * @param {string} password - User password
 */
async function registerUser(page, email, name, password) {
  await page.goto('http://localhost:8082/register');
  await page.fill('input[name="email"]', email);
  await page.fill('input[name="name"]', name);
  await page.fill('input[name="password"]', password);
  await page.fill('input[name="confirmPassword"]', password);
  await page.click('button[type="submit"]');

  // Wait for navigation to complete
  await page.waitForLoadState('networkidle');

  // Check if registration was successful
  const url = page.url();
  if (!url.includes('/dashboard') && !url.includes('/login')) {
    throw new Error(`Registration failed. Current URL: ${url}`);
  }
}

/**
 * Logout from the application
 * @param {import('@playwright/test').Page} page - Playwright page
 */
async function logout(page) {
  console.log('Starting logout process...');

  // Take screenshot before logout
  await page.screenshot({ path: 'before-logout.png' });

  // Click on the dropdown toggle (user email)
  console.log('Clicking on dropdown toggle...');
  await page.click('.dropdown-toggle');

  // Take screenshot after clicking dropdown
  await page.screenshot({ path: 'dropdown-open.png' });

  // Click on the logout link
  console.log('Clicking on logout link...');
  await page.click('a[href="/logout"]');

  // Wait for logout to complete
  console.log('Waiting for logout to complete...');
  await page.waitForLoadState('networkidle');

  // Take screenshot after logout
  await page.screenshot({ path: 'after-logout.png' });

  // Verify we're logged out
  const currentUrl = page.url();
  console.log(`Current URL after logout: ${currentUrl}`);

  if (!currentUrl.includes('/')) {
    console.log('Logout failed. Checking page content...');
    const pageContent = await page.textContent('body');
    console.log('Page content:', pageContent.substring(0, 500) + '...');
    throw new Error(`Logout failed. Current URL: ${currentUrl}`);
  }

  console.log('Logout successful.');
}

module.exports = {
  TEST_USER,
  login,
  registerUser,
  logout
};
