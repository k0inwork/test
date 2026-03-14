const { test, expect } = require('@playwright/test');

test('Capture console logs during project creation and edit view', async ({ page }) => {
  // Capture console logs
  page.on('console', msg => {
    console.log(`BROWSER CONSOLE [${msg.type()}]: ${msg.text()}`);
  });

  // Capture page errors
  page.on('pageerror', err => {
    console.log(`BROWSER ERROR: ${err.message}`);
  });

  // Navigate to Phase 2 dev environment
  await page.goto('http://localhost:8788/dashboard');

  // Verify dashboard loads (the redirect from / happened, auth bypassed)
  await expect(page).toHaveURL(/.*\/dashboard.*/);

  // Click on "Create Project"
  await page.locator('#create-btn').click();

  // Wait for the modal dialog to appear
  await page.waitForSelector('#project-dialog-title');
  await expect(page.locator('#project-dialog-title')).toHaveText('New Project');

  // Type a project name
  await page.fill('#project-name', 'test-log-capture-project');

  // Wait for debounce
  await page.waitForTimeout(1000);

  // Verify that the error messages do not appear
  await expect(page.locator('#project-name-error')).toHaveText('');

  // The Create button should be enabled
  const btn = page.locator('#project-submit');
  await expect(btn).not.toBeDisabled();

  // Create the project
  await btn.click();

  // Wait for the redirect to the edit page
  await page.waitForURL(/.*\/edit\/test-log-capture-project.*/);

  console.log("Navigated to /edit view. Waiting 10 seconds for Wanix runtime to initialize and print logs...");

  // Wait for Wanix to initialize and potentially panic
  await page.waitForTimeout(10000);
});
