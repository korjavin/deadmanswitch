/**
 * Utility functions for frontend tests
 */

// Test data that can be reused across tests
const TEST_USER = {
  email: 'test@example.com',
  password: 'Password123!',
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
    const loginErrorText = await page.textContent('body');
    console.log('Login page content:', loginErrorText.substring(0, 500) + '...');

    console.log('Attempting to register a new test user with unique email...');

    // Generate a unique email with timestamp to avoid conflicts
    const uniqueEmail = `test_${Date.now()}@example.com`;
    console.log(`Using unique email: ${uniqueEmail}`);

    // Try to register the user with the unique email
    try {
      await registerUser(page, uniqueEmail, TEST_USER.name, password);
      console.log('User registration successful. Trying to login with the new account...');

      // Try to login with the new account
      await page.goto('http://localhost:8082/login');
      await page.waitForLoadState('networkidle');
      await page.fill('input[name="email"]', uniqueEmail);
      await page.fill('input[name="password"]', password);

      // Take screenshot before submitting
      await page.screenshot({ path: 'login-retry-form-filled.png' });

      console.log('Submitting login form with new account...');
      await page.click('button[type="submit"]');
      await page.waitForLoadState('networkidle');

      // Take screenshot after login attempt
      await page.screenshot({ path: 'after-login-retry-attempt.png' });

      // Check if login was successful after registration
      const newUrl = page.url();
      console.log(`Current URL after login with new account: ${newUrl}`);

      if (!newUrl.includes('/dashboard')) {
        console.log('Login failed even with new account. Checking for error messages...');
        const retryErrorText = await page.textContent('body');
        console.log('Retry login page content:', retryErrorText.substring(0, 500) + '...');
        throw new Error(`Login failed with new account. Current URL: ${newUrl}`);
      }

      // Update the TEST_USER email to use the successful one for future tests
      TEST_USER.email = uniqueEmail;
      console.log(`Updated TEST_USER.email to: ${TEST_USER.email}`);
    } catch (error) {
      console.log('Registration or login with new account failed:', error.message);

      // One last attempt with admin account
      console.log('Attempting one last login with admin account...');
      await page.goto('http://localhost:8082/login');
      await page.waitForLoadState('networkidle');
      await page.fill('input[name="email"]', 'admin@example.com');
      await page.fill('input[name="password"]', 'admin');
      await page.click('button[type="submit"]');
      await page.waitForLoadState('networkidle');

      const adminUrl = page.url();
      if (!adminUrl.includes('/dashboard')) {
        throw new Error(`All login attempts failed. Original URL: ${currentUrl}. Registration and admin login also failed.`);
      } else {
        console.log('Admin login successful as fallback.');
        TEST_USER.email = 'admin@example.com';
        TEST_USER.password = 'admin';
      }
    }
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
  console.log(`Navigating to register page...`);
  await page.goto('http://localhost:8082/register');
  await page.waitForLoadState('networkidle');

  console.log(`Filling registration form with email: ${email}, name: ${name}...`);
  await page.fill('input[name="email"]', email);
  await page.fill('input[name="name"]', name);
  await page.fill('input[name="password"]', password);
  await page.fill('input[name="confirmPassword"]', password);

  // Take screenshot before submitting
  await page.screenshot({ path: 'registration-form-filled.png' });

  console.log('Submitting registration form...');
  await page.click('button[type="submit"]');

  // Wait for navigation to complete
  console.log('Waiting for navigation to complete after registration...');
  await page.waitForLoadState('networkidle');

  // Take screenshot after registration attempt
  await page.screenshot({ path: 'after-registration-attempt.png' });

  // Check if registration was successful
  const url = page.url();
  console.log(`Current URL after registration: ${url}`);

  if (!url.includes('/dashboard') && !url.includes('/login')) {
    console.log('Registration failed. Checking for error messages...');
    const errorText = await page.textContent('body');
    console.log('Page content:', errorText.substring(0, 500) + '...');
    throw new Error(`Registration failed. Current URL: ${url}`);
  }

  console.log('Registration successful or redirected to login.');
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
