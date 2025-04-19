/**
 * Utility functions for frontend tests
 */

/**
 * Login to the application
 * @param {import('@playwright/test').Page} page - Playwright page
 * @param {string} email - User email
 * @param {string} password - User password
 */
async function login(page, email, password) {
  await page.goto('/login');
  await page.fill('input[name="email"]', email);
  await page.fill('input[name="password"]', password);
  await page.click('button[type="submit"]');
  
  // Wait for navigation to complete
  await page.waitForURL('/dashboard');
}

/**
 * Register a new user
 * @param {import('@playwright/test').Page} page - Playwright page
 * @param {string} email - User email
 * @param {string} password - User password
 */
async function registerUser(page, email, password) {
  await page.goto('/register');
  await page.fill('input[name="email"]', email);
  await page.fill('input[name="password"]', password);
  await page.fill('input[name="password_confirm"]', password);
  await page.click('button[type="submit"]');
  
  // Wait for registration to complete
  await page.waitForURL('/login');
}

/**
 * Logout from the application
 * @param {import('@playwright/test').Page} page - Playwright page
 */
async function logout(page) {
  // Click on the dropdown toggle (user email)
  await page.click('.dropdown-toggle');
  
  // Click on the logout link
  await page.click('a[href="/logout"]');
  
  // Wait for logout to complete
  await page.waitForURL('/');
}

module.exports = {
  login,
  registerUser,
  logout
};
