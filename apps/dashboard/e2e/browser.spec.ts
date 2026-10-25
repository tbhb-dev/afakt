import { env } from "node:process";
import { expect, test } from "@playwright/test";

// mise.toml pins the browser build that the installed Playwright release runs.
const pinned: string | undefined = env["AFAKT_WEB_CHROMIUM"];

test("the browser matches the pinned version", ({ browser }) => {
  expect(browser.version()).toBe(pinned);
});
