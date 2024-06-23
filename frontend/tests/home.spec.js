// @ts-check
import { test, expect } from '@playwright/test';

/**
 * @param {import('@playwright/test').Page} page
 */
async function setupEndpoints(page) {
  await page.route('**/api/projects', async route => {
    const json = JSON.parse(`[
  {
    "id": "1",
    "title": "sharp-cooking-web",
    "issues": null,
    "statuses": null,
    "iterations": null
  },
  {
    "id": "2",
    "title": "Agile Project",
    "issues": null,
    "statuses": null,
    "iterations": null
  }
]`);
    await route.fulfill({ json });
  });

  await page.route('**/api/projects/2/iterations', async route => {
    const json = JSON.parse(`[
  {
    "id": "1",
    "title": "Iteration",
    "startDate": "2024-06-18T00:00:00Z",
    "endDate": "2024-07-02T00:00:00Z"
  },
  {
    "id": "2",
    "title": "Iteration 2",
    "startDate": "2024-07-02T00:00:00Z",
    "endDate": "2024-07-16T00:00:00Z"
  }
]`);
    await route.fulfill({ json });
  });

  await page.route('**/api/projects/2/burnup', async route => {
    const json = JSON.parse(`[
  { "status": "Approved", "projectDay": "2024-05-22T00:00:00Z", "qty": 1 },
  { "status": "New", "projectDay": "2024-05-22T00:00:00Z", "qty": 2 },
  { "status": "InProgress", "projectDay": "2024-05-22T00:00:00Z", "qty": 3 }
]`);
    await route.fulfill({ json });
  });

  
  await page.route('**/api/projects/2/iterations/1/burndown', async route => {
    const json = JSON.parse(`[
  { "iterationDay": "2024-06-18T00:00:00Z", "remaining": 10, "ideal": 10 },
  { "iterationDay": "2024-06-19T00:00:00Z", "remaining": 8, "ideal": 5 },
  { "iterationDay": "2024-06-20T00:00:00Z", "remaining": 6, "ideal": 0 }
]`);
    await route.fulfill({ json });
  });
}

test.beforeEach(async ({ page }) => {
  await setupEndpoints(page);
});

test('has title', async ({ page }) => {
  await page.goto('/');

  await expect(page).toHaveTitle(/github-charts/);
});

test('loads projects', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('combobox').first().getByRole('option')).toHaveCount(3);
});

test('loads iterations on project selection', async ({ page }) => {
  await page.goto('/');

  await page.getByRole('combobox').first().selectOption('2');
  await expect(page.getByRole('combobox').nth(1).getByRole('option')).toHaveCount(3);
});

test('loads charts', async ({ page }) => {
  await page.goto('/');
  
  await page.getByRole('combobox').first().selectOption('2');
  await page.getByRole('combobox').nth(1).selectOption('1');

  await expect(page.locator('.card').first()).toBeVisible();
  await expect(page.locator('.card').nth(1)).toBeVisible();
});
