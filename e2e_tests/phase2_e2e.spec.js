const { test, expect } = require('@playwright/test');

test('Phase 2 mock duplicate project check should work', async ({ page }) => {
  // Navigate to Phase 2 dev environment
  await page.goto('http://localhost:8788/dashboard');

  // Verify dashboard loads (the redirect from / happened, auth bypassed)
  await expect(page).toHaveURL(/.*\/dashboard.*/);

  // Click on "New Project"
  await page.locator('text=New Project').first().click();

  // Wait for the modal dialog to appear
  await page.waitForSelector('#project-dialog-title');
  await expect(page.locator('#project-dialog-title')).toHaveText('New Project');

  // Type a non-existent project name
  await page.fill('#project-name', 'test-new-project-does-not-exist');

  // Wait for debounce
  await page.waitForTimeout(1000);

  // Verify that the error messages do not appear
  await expect(page.locator('#project-name-error')).toHaveText('');

  // The Create button should be enabled
  const btn = page.locator('#project-submit');
  await expect(btn).not.toBeDisabled();
});
