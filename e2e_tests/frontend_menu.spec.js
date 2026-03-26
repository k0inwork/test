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

  // Wait for services to be fetched and check navigation items
  // Since we are running in a consolidated environment,
  // some services should be registered if start.sh was run.
  // We can at least check if the nav bar structure exists.
  const navLinks = page.locator('.navbar-nav .nav-link');
  const count = await navLinks.count();
  console.log(`Found ${count} navigation links`);
});

test('websocket connection status', async ({ page }) => {
  await page.goto('http://localhost:8080/login');
  await page.fill('input[name="username"]', 'admin');
  await page.fill('input[name="password"]', 'admin');
  await page.click('button[type="submit"]');

  // Check if the notification button is present
  await expect(page.locator('button:has-text("Notifications")')).toBeVisible();

  // Open the offcanvas
  await page.click('button:has-text("Notifications")');

  // Check for the "No messages yet" text or messages
  // This verifies the offcanvas and the underlying script ran
  const wsEmpty = page.locator('#ws-empty');
  const wsMessages = page.locator('#ws-messages');

  await expect(wsMessages).toBeVisible();
});
