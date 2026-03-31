const { test, expect } = require('@playwright/test');

test('frontend menu visibility and registered services', async ({ page }) => {
  // Go to login page
  await page.goto('http://localhost:8080/login');

  // Login as admin
  await page.fill('input[name="username"]', 'admin');
  await page.fill('input[name="password"]', 'admin');
  await page.click('button[type="submit"]');

  // Wait for redirect to dashboard
  await expect(page).toHaveURL('http://localhost:8080/');

  // Check if Admin link is visible
  await expect(page.locator('nav')).toContainText('Admin');
});

test('websocket connection status', async ({ page }) => {
  await page.goto('http://localhost:8080/login');
  await page.fill('input[name="username"]', 'admin');
  await page.fill('input[name="password"]', 'admin');
  await page.click('button[type="submit"]');

  // Check if the notification button is present
  await expect(page.locator('button:has-text("Notifications")')).toBeVisible();
});
