import { expect, test } from "@playwright/test";
import { expectAccessible, recordErrors, recordRequests } from "./checks.ts";

test("the blank page loads without errors", async ({ page }) => {
  const errors = recordErrors(page);
  const requests = recordRequests(page);
  await page.goto("./");
  await expect(page).toHaveTitle("afakt");
  await expect(page.locator("#root > main")).toBeEmpty();
  expect(errors).toStrictEqual([]);
  // Every request goes to the server that served the page.
  const { origin } = new URL(page.url());
  expect(
    requests.filter((url) => new URL(url).origin !== origin),
  ).toStrictEqual([]);
});

test("the blank page passes the accessibility scan", async ({ page }) => {
  await page.goto("./");
  await expect(page.locator("#root > main")).toBeAttached();
  await expectAccessible(page);
});
