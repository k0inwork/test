const { test, expect } = require('@playwright/test');

test('Phase 2 complete project creation flow', async ({ page }) => {
  // Navigate to Phase 2 dev environment root, which redirects to signin or dashboard
  await page.goto('http://localhost:8788/');

  // Should redirect to dashboard in mock mode
  await expect(page).toHaveURL(/.*\/dashboard.*/);

  // Ensure grid is visible
  await page.waitForSelector('#projects-grid');

  // Click on "Create Project"
  await page.locator('#create-btn').click();

  // Wait for the modal dialog to appear
  await page.waitForSelector('#project-dialog-title');
  await expect(page.locator('#project-dialog-title')).toHaveText('New Project');

  // Type a project name
  const projName = 'my-test-project-' + Date.now();
  await page.fill('#project-name', projName);
  await page.fill('#project-desc', 'This is a test project');

  // Wait for debounce and duplicate check
  await page.waitForTimeout(1000);

  // The Create button should be enabled
  const btn = page.locator('#project-submit');
  await expect(btn).not.toBeDisabled();

  // Create the project
  await btn.click();

  // Wait for the redirect to the edit page which indicates success
  await page.waitForURL(new RegExp(`.*/edit/${projName}\\?user=admin.*`));

  // The page should contain the iframe and not "Not Found"
  const bodyText = await page.locator('body').innerText();
  expect(bodyText).not.toContain('Not Found');

  // Wait for the iframe to be present (the editor)
  const iframe = page.locator('iframe');
  await expect(iframe).toBeVisible();

  // Test the PUM bundle is mounted in the terminal
  // The terminal canvas is inside the apptron runtime
  await page.waitForTimeout(5000);
  await page.keyboard.type('ls\n');
  await page.waitForTimeout(1000);
});
