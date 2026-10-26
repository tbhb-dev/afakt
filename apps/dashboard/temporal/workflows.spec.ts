import { expect, test } from "@playwright/test";
import { recordErrors } from "../e2e/checks.ts";

const httpError = 400;
const workflowList = /\/temporal\/namespaces\/default\/workflows/u;

test("the embedded namespace lists a workflow with no browser failures", async ({
  page,
}) => {
  await page.clock.install();
  const errors = recordErrors(page);
  const failed: string[] = [];
  const requests: string[] = [];
  page.on("request", (request) => requests.push(request.url()));
  page.on("requestfailed", (request) => failed.push(request.url()));
  page.on("response", (result) => {
    if (result.status() >= httpError) {
      failed.push(`${String(result.status())} ${result.url()}`);
    }
  });

  await page.goto("/temporal/namespaces");
  await page.getByRole("link", { name: "default", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Namespace: default", exact: true }),
  ).toBeVisible();
  await page.getByRole("link", { name: "Workflows", exact: true }).click();
  await expect(page).toHaveURL(workflowList);

  const workflowId = `ui-proof-${crypto.randomUUID()}`;
  const cookies = await page.context().cookies();
  const csrf = cookies.find((cookie) => cookie.name === "_csrf");
  const response = await page.request.post(
    `/temporal/api/v1/namespaces/default/workflows/${workflowId}`,
    {
      headers: { "X-CSRF-Token": csrf?.value ?? "" },
      data: {
        workflowId,
        workflowType: { name: "ui-proof" },
        taskQueue: { name: "ui-proof" },
        requestId: crypto.randomUUID(),
        workflowExecutionTimeout: "3600s",
      },
    },
  );
  expect(response.ok(), await response.text()).toBe(true);
  await expect(async () => {
    await page.reload();
    await expect(
      page.getByRole("link", { name: workflowId, exact: true }),
    ).toBeVisible();
  }).toPass({ timeout: 30_000 });

  expect(
    requests.some((url) => url.includes("/temporal/_app/immutable/")),
  ).toBe(true);
  expect(
    requests.every((url) => url.startsWith("http://127.0.0.1:4186/temporal/")),
  ).toBe(true);
  expect(failed).toStrictEqual([]);
  expect(errors).toStrictEqual([]);
});
