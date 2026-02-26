import { expect, test } from "@playwright/test";

test("inventory loads without 500 fallback", async ({ page, request }) => {
  const itemsResp = await request.get("/api/items");
  expect(itemsResp.ok()).toBeTruthy();
  const payload = await itemsResp.json();
  expect(Array.isArray(payload.items)).toBeTruthy();

  await page.goto("/inventory");

  await expect(page.getByText("Oops! Something went wrong")).toHaveCount(0);
  await expect(page.getByRole("heading", { name: "Inventory" })).toBeVisible();
  await expect(page.locator("main")).toBeVisible();
});

