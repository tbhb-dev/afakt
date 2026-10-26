import { expect, test } from "@playwright/test";
import { recordErrors, recordRequests } from "../e2e/checks.ts";

const httpOk = 200;

test("all targets serve the dashboard, namespace and healthy probes", async ({
  page,
  request,
}) => {
  await page.clock.install();
  const errors = recordErrors(page);
  const requests = recordRequests(page);
  const dashboard = await page.goto("/dashboard/");
  expect(dashboard?.status()).toBe(httpOk);
  await expect(page).toHaveTitle("afakt");
  await expect(page.locator("#root > main")).toBeEmpty();
  await page.goto("/temporal/namespaces");
  await expect(
    page.getByRole("link", { name: "default", exact: true }),
  ).toBeVisible();
  const responses = await Promise.all(
    ["/livez", "/readyz", "/api/v1/"].map((path) => request.get(path)),
  );
  for (const response of responses) {
    expect(response.status()).toBe(httpOk);
  }
  expect(
    requests.every((url) => url.startsWith("http://127.0.0.1:8080/")),
  ).toBe(true);
  expect(errors).toStrictEqual([]);
});
