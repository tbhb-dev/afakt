import { env } from "node:process";
import { expect, test } from "@playwright/test";

// mise.toml pins the browser build that the installed Playwright release runs.
const { AFAKT_WEB_CHROMIUM: pinned }: Record<string, string | undefined> = env;

test("the browser matches the pinned version", ({ browser }) => {
  expect(browser.version()).toBe(pinned);
});
