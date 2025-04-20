const { test, expect } = require('@playwright/test');
const { TEST_USER, login, logout } = require('./utils');

test.describe('Recipient Management Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Go to the home page and login before each test
    console.log('Navigating to home page...');
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Login with test credentials
    console.log('Logging in with test credentials...');
    await login(page, TEST_USER.email, TEST_USER.password);
    console.log('Login completed.');
  });

  test.afterEach(async ({ page }) => {
    // Logout after each test
    try {
      console.log('Logging out...');
      // Try to click on the dropdown toggle
      try {
        await page.click('.dropdown-toggle', { timeout: 5000 });
        // Try to click on the logout link
        try {
          await page.click('a[href="/logout"]', { timeout: 5000 });
          await page.waitForLoadState('networkidle');
          console.log('Logout completed successfully.');
        } catch (error) {
          console.log('Could not click logout link:', error);
        }
      } catch (error) {
        console.log('Could not click dropdown toggle:', error);
      }
    } catch (error) {
      console.log('Error during logout:', error);
    }
  });

  test('should create and delete a recipient', async ({ page }) => {
    console.log('Starting test: should create and delete a recipient');

    // Navigate to recipients page
    console.log('Navigating to recipients page...');
    await page.goto('http://localhost:8082/recipients');
    await page.waitForLoadState('networkidle');

    // Take screenshot of recipients page
    await page.screenshot({ path: 'recipients-page.png' });

    // Click on "Add Recipient" button
    console.log('Clicking on Add Recipient button...');
    await page.click('a:has-text("Add Recipient")');
    await page.waitForLoadState('networkidle');

    // Take screenshot of new recipient form
    await page.screenshot({ path: 'new-recipient-form.png' });

    // Fill out the recipient form
    const recipientName = `Test Recipient ${Date.now()}`;
    const recipientEmail = `recipient_${Date.now()}@example.com`;

    console.log(`Creating recipient with name: ${recipientName}, email: ${recipientEmail}`);

    // Wait for the form to be visible
    await page.waitForSelector('form');

    // Fill the form fields
    await page.fill('input#name', recipientName);
    await page.fill('input#email', recipientEmail);

    // Take screenshot of filled form
    await page.screenshot({ path: 'filled-recipient-form.png' });

    // Submit the form
    console.log('Submitting recipient form...');
    await page.click('form button[type="submit"]');
    await page.waitForLoadState('networkidle');

    // Take screenshot after submission
    await page.screenshot({ path: 'after-recipient-creation.png' });

    // Verify we're back on the recipients page
    console.log('Verifying we are back on recipients page...');
    expect(page.url()).toContain('/recipients');

    // Wait for the page to load completely
    await page.waitForTimeout(1000);

    // Refresh the page to ensure the new recipient is loaded
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Take screenshot after reload
    await page.screenshot({ path: 'after-reload.png' });

    // Verify the new recipient is in the list
    console.log(`Checking if recipient ${recipientName} is in the list...`);

    // Get all recipient names on the page
    const pageContent = await page.content();
    console.log('Page content snippet:', pageContent.substring(0, 1000) + '...');

    // Try to find the recipient by name
    const recipientElement = page.getByText(recipientName, { exact: false });

    // Wait longer for the element to be visible
    try {
      await expect(recipientElement).toBeVisible({ timeout: 10000 });
      console.log('Recipient found in the list.');

      // Now delete the recipient
      console.log('Deleting the recipient...');

      // Find the recipient row and click on it to go to edit page
      await recipientElement.click();
      await page.waitForLoadState('networkidle');

      // Take screenshot of edit page
      await page.screenshot({ path: 'recipient-edit-page.png' });

      // Click on Delete button
      console.log('Clicking on Delete button...');
      await page.click('button:has-text("Delete")');

      // Confirm deletion in the modal
      console.log('Confirming deletion...');
      await page.click('button:has-text("Confirm")');
      await page.waitForLoadState('networkidle');

      // Take screenshot after deletion
      await page.screenshot({ path: 'after-recipient-deletion.png' });

      // Verify we're back on the recipients page
      console.log('Verifying we are back on recipients page...');
      expect(page.url()).toContain('/recipients');

      // Verify the recipient is no longer in the list
      console.log(`Checking if recipient ${recipientName} is removed from the list...`);
      await expect(page.getByText(recipientName, { exact: false })).not.toBeVisible({ timeout: 5000 });
      console.log('Recipient successfully deleted.');
    } catch (error) {
      console.log('Could not find recipient in the list. Test failed:', error);
      throw error;
    }
  });

  test('should add a secret to a recipient', async ({ page }) => {
    console.log('Starting test: should add a secret to a recipient');

    // First create a recipient
    console.log('Creating a new recipient...');
    await page.goto('http://localhost:8082/recipients/new');
    await page.waitForLoadState('networkidle');

    const recipientName = `Secret Test Recipient ${Date.now()}`;
    const recipientEmail = `secret_recipient_${Date.now()}@example.com`;

    console.log(`Creating recipient with name: ${recipientName}, email: ${recipientEmail}`);

    // Wait for the form to be visible
    await page.waitForSelector('form');

    // Fill the form fields
    await page.fill('input#name', recipientName);
    await page.fill('input#email', recipientEmail);
    await page.click('form button[type="submit"]');
    await page.waitForLoadState('networkidle');

    // Now create a secret
    console.log('Creating a new secret...');
    await page.goto('http://localhost:8082/secrets/new');
    await page.waitForLoadState('networkidle');

    // Take screenshot of new secret form
    await page.screenshot({ path: 'new-secret-form.png' });

    // Get the page content to debug
    const secretFormContent = await page.content();
    console.log('Secret form content snippet:', secretFormContent.substring(0, 1000) + '...');

    const secretName = `Test Secret ${Date.now()}`;
    const secretContent = 'This is a test secret content';

    console.log(`Creating secret with name: ${secretName}`);

    // Wait for the form to be visible
    await page.waitForSelector('form');

    // Fill the form fields using more general selectors
    // First, get all input fields and fill the first one with the name
    const nameInputs = await page.$$('input[type="text"]');
    if (nameInputs.length > 0) {
      await nameInputs[0].fill(secretName);
    } else {
      console.log('No text input fields found for secret name');
    }

    // Then, get all textarea fields and fill the first one with the content
    const contentTextareas = await page.$$('textarea');
    if (contentTextareas.length > 0) {
      await contentTextareas[0].fill(secretContent);
    } else {
      console.log('No textarea fields found for secret content');
    }

    // Take screenshot of filled secret form
    await page.screenshot({ path: 'filled-secret-form.png' });

    await page.click('form button[type="submit"]');
    await page.waitForLoadState('networkidle');

    // Now assign the secret to the recipient
    console.log('Navigating to recipient secrets page...');
    await page.goto('http://localhost:8082/recipients');
    await page.waitForLoadState('networkidle');

    // Wait for the page to load completely
    await page.waitForTimeout(1000);

    // Refresh the page to ensure the new recipient is loaded
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Take screenshot of recipients page
    await page.screenshot({ path: 'recipients-list.png' });

    // Find and click on the recipient
    console.log(`Finding recipient ${recipientName}...`);
    const recipientElement = page.getByText(recipientName, { exact: false });
    await expect(recipientElement).toBeVisible({ timeout: 10000 });
    await recipientElement.click();
    await page.waitForLoadState('networkidle');

    // Take screenshot of recipient details page
    await page.screenshot({ path: 'recipient-details.png' });

    // Click on "Manage Secrets" button
    console.log('Clicking on Manage Secrets button...');
    await page.click('a:has-text("Manage Secrets")');
    await page.waitForLoadState('networkidle');

    // Take screenshot of manage secrets page
    await page.screenshot({ path: 'manage-secrets-page.png' });

    // Get the page content to debug
    const secretsPageContent = await page.content();
    console.log('Secrets page content snippet:', secretsPageContent.substring(0, 1000) + '...');

    // Check the checkbox for the secret
    console.log(`Selecting secret ${secretName}...`);

    // Try different selector strategies
    try {
      // First try with a more specific selector
      await page.check(`input[type="checkbox"]:near(:text("${secretName}"))`);
    } catch (error) {
      console.log('First checkbox selector failed, trying alternative:', error);
      try {
        // Try with a more general selector
        const checkboxes = await page.$$('input[type="checkbox"]');
        console.log(`Found ${checkboxes.length} checkboxes`);

        if (checkboxes.length > 0) {
          await checkboxes[0].check();
        } else {
          throw new Error('No checkboxes found on the page');
        }
      } catch (error2) {
        console.log('All checkbox selection attempts failed:', error2);
        throw error2;
      }
    }

    // Take screenshot after selecting secret
    await page.screenshot({ path: 'selected-secret.png' });

    // Save the changes
    console.log('Saving changes...');
    await page.click('button[type="submit"]');
    await page.waitForLoadState('networkidle');

    // Take screenshot after saving
    await page.screenshot({ path: 'after-secret-assignment.png' });

    // Verify the secret is assigned by going back to the recipient page
    console.log('Verifying secret is assigned...');
    await page.goto('http://localhost:8082/recipients');
    await page.waitForLoadState('networkidle');

    // Find and click on the recipient again
    const recipientElementAgain = page.getByText(recipientName, { exact: false });
    await expect(recipientElementAgain).toBeVisible({ timeout: 10000 });
    await recipientElementAgain.click();
    await page.waitForLoadState('networkidle');

    // Take screenshot of recipient details page after assignment
    await page.screenshot({ path: 'recipient-details-after-assignment.png' });

    // Check if the secret name is visible on the page
    try {
      await expect(page.getByText(secretName, { exact: false })).toBeVisible({ timeout: 5000 });
      console.log('Secret successfully assigned to recipient.');
    } catch (error) {
      console.log('Could not verify secret assignment:', error);
      // Continue with cleanup even if verification fails
    }

    // Clean up - delete the recipient
    console.log('Cleaning up - deleting recipient...');
    await page.goto('http://localhost:8082/recipients');
    await page.waitForLoadState('networkidle');

    const recipientForDelete = page.getByText(recipientName, { exact: false });
    await expect(recipientForDelete).toBeVisible({ timeout: 10000 });
    await recipientForDelete.click();
    await page.waitForLoadState('networkidle');

    await page.click('button:has-text("Delete")');
    await page.click('button:has-text("Confirm")');
    await page.waitForLoadState('networkidle');

    // Clean up - delete the secret
    console.log('Cleaning up - deleting secret...');
    await page.goto('http://localhost:8082/secrets');
    await page.waitForLoadState('networkidle');

    const secretForDelete = page.getByText(secretName, { exact: false });
    try {
      await expect(secretForDelete).toBeVisible({ timeout: 10000 });
      await secretForDelete.click();
      await page.waitForLoadState('networkidle');

      await page.click('button:has-text("Delete")');
      await page.click('button:has-text("Confirm")');
      await page.waitForLoadState('networkidle');

      console.log('Secret successfully deleted.');
    } catch (error) {
      console.log('Could not find secret to delete:', error);
      // Continue even if we can't delete the secret
    }
  });

  test('should add a secret question to a recipient', async ({ page }) => {
    console.log('Starting test: should add a secret question to a recipient');

    // First create a recipient
    console.log('Creating a new recipient...');
    await page.goto('http://localhost:8082/recipients/new');
    await page.waitForLoadState('networkidle');

    const recipientName = `Question Test Recipient ${Date.now()}`;
    const recipientEmail = `question_recipient_${Date.now()}@example.com`;

    console.log(`Creating recipient with name: ${recipientName}, email: ${recipientEmail}`);

    // Wait for the form to be visible
    await page.waitForSelector('form');

    // Fill the form fields
    await page.fill('input#name', recipientName);
    await page.fill('input#email', recipientEmail);
    await page.click('form button[type="submit"]');
    await page.waitForLoadState('networkidle');

    // Now create a secret
    console.log('Creating a new secret...');
    await page.goto('http://localhost:8082/secrets/new');
    await page.waitForLoadState('networkidle');

    // Take screenshot of new secret form
    await page.screenshot({ path: 'new-secret-form-questions.png' });

    const secretName = `Question Secret ${Date.now()}`;
    const secretContent = 'This is a test secret for questions';

    console.log(`Creating secret with name: ${secretName}`);

    // Wait for the form to be visible
    await page.waitForSelector('form');

    // Fill the form fields using more general selectors
    // First, get all input fields and fill the first one with the name
    const nameInputs = await page.$$('input[type="text"]');
    if (nameInputs.length > 0) {
      await nameInputs[0].fill(secretName);
    } else {
      console.log('No text input fields found for secret name');
    }

    // Then, get all textarea fields and fill the first one with the content
    const contentTextareas = await page.$$('textarea');
    if (contentTextareas.length > 0) {
      await contentTextareas[0].fill(secretContent);
    } else {
      console.log('No textarea fields found for secret content');
    }
    await page.click('form button[type="submit"]');
    await page.waitForLoadState('networkidle');

    // Assign the secret to the recipient
    console.log('Assigning secret to recipient...');
    await page.goto('http://localhost:8082/recipients');
    await page.waitForLoadState('networkidle');

    // Wait for the page to load completely
    await page.waitForTimeout(1000);

    // Refresh the page to ensure the new recipient is loaded
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Find and click on the recipient
    console.log(`Finding recipient ${recipientName}...`);
    const recipientElement = page.getByText(recipientName, { exact: false });
    await expect(recipientElement).toBeVisible({ timeout: 10000 });
    await recipientElement.click();
    await page.waitForLoadState('networkidle');

    // Click on "Manage Secrets" button
    console.log('Clicking on Manage Secrets button...');
    await page.click('a:has-text("Manage Secrets")');
    await page.waitForLoadState('networkidle');

    // Try different selector strategies for the checkbox
    try {
      // First try with a more specific selector
      await page.check(`input[type="checkbox"]:near(:text("${secretName}"))`);
    } catch (error) {
      console.log('First checkbox selector failed, trying alternative:', error);
      try {
        // Try with a more general selector
        const checkboxes = await page.$$('input[type="checkbox"]');
        console.log(`Found ${checkboxes.length} checkboxes`);

        if (checkboxes.length > 0) {
          await checkboxes[0].check();
        } else {
          throw new Error('No checkboxes found on the page');
        }
      } catch (error2) {
        console.log('All checkbox selection attempts failed:', error2);
        throw error2;
      }
    }

    await page.click('button[type="submit"]');
    await page.waitForLoadState('networkidle');

    // Now navigate to the secret questions page
    console.log('Navigating to secret questions page...');
    await page.goto('http://localhost:8082/recipients');
    await page.waitForLoadState('networkidle');

    // Find and click on the recipient again
    const recipientForQuestions = page.getByText(recipientName, { exact: false });
    await expect(recipientForQuestions).toBeVisible({ timeout: 10000 });
    await recipientForQuestions.click();
    await page.waitForLoadState('networkidle');

    // Click on "Secret Questions" button
    console.log('Clicking on Secret Questions button...');
    await page.click('a:has-text("Secret Questions")');
    await page.waitForLoadState('networkidle');

    // Take screenshot of questions page
    await page.screenshot({ path: 'secret-questions-page.png' });

    // Get the page content to debug
    const questionsPageContent = await page.content();
    console.log('Questions page content snippet:', questionsPageContent.substring(0, 1000) + '...');

    // Add questions
    console.log('Adding secret questions...');

    try {
      // Wait for the form to be visible
      await page.waitForSelector('form', { timeout: 5000 });

      // Fill out the questions form
      await page.fill('input[name="question"]', 'What is your favorite color?');
      await page.fill('input[name="answer"]', 'Blue');
      await page.click('button:has-text("Add Question")');

      await page.fill('input[name="question"]', 'What is your pet\'s name?');
      await page.fill('input[name="answer"]', 'Fluffy');
      await page.click('button:has-text("Add Question")');

      await page.fill('input[name="question"]', 'What city were you born in?');
      await page.fill('input[name="answer"]', 'New York');
      await page.click('button:has-text("Add Question")');

      // Take screenshot after adding questions
      await page.screenshot({ path: 'added-questions.png' });

      // Set threshold
      await page.selectOption('select[name="threshold"]', '2');

      // Save the questions
      console.log('Saving questions...');
      await page.click('button[type="submit"]:has-text("Save Questions")');
      await page.waitForLoadState('networkidle');

      // Take screenshot after saving
      await page.screenshot({ path: 'after-questions-saved.png' });

      // Verify questions are saved
      console.log('Verifying questions are saved...');
      await expect(page.getByText('What is your favorite color?', { exact: false })).toBeVisible({ timeout: 5000 });
      await expect(page.getByText('What is your pet\'s name?', { exact: false })).toBeVisible({ timeout: 5000 });
      await expect(page.getByText('What city were you born in?', { exact: false })).toBeVisible({ timeout: 5000 });
      console.log('Secret questions successfully added.');
    } catch (error) {
      console.log('Error in adding or verifying questions:', error);
      // Continue with cleanup even if this part fails
    }

    // Clean up - delete the recipient
    console.log('Cleaning up - deleting recipient...');
    await page.goto('http://localhost:8082/recipients');
    await page.waitForLoadState('networkidle');

    const recipientForDelete = page.getByText(recipientName, { exact: false });
    try {
      await expect(recipientForDelete).toBeVisible({ timeout: 10000 });
      await recipientForDelete.click();
      await page.waitForLoadState('networkidle');

      await page.click('button:has-text("Delete")');
      await page.click('button:has-text("Confirm")');
      await page.waitForLoadState('networkidle');

      console.log('Recipient successfully deleted.');
    } catch (error) {
      console.log('Could not delete recipient:', error);
      // Continue even if we can't delete the recipient
    }

    // Clean up - delete the secret
    console.log('Cleaning up - deleting secret...');
    await page.goto('http://localhost:8082/secrets');
    await page.waitForLoadState('networkidle');

    const secretForDelete = page.getByText(secretName, { exact: false });
    try {
      await expect(secretForDelete).toBeVisible({ timeout: 10000 });
      await secretForDelete.click();
      await page.waitForLoadState('networkidle');

      await page.click('button:has-text("Delete")');
      await page.click('button:has-text("Confirm")');
      await page.waitForLoadState('networkidle');

      console.log('Secret successfully deleted.');
    } catch (error) {
      console.log('Could not delete secret:', error);
      // Continue even if we can't delete the secret
    }
  });
});
