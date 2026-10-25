import { act } from "react";
import { expect, test } from "vitest";

test("the entry point mounts the blank page into #root", async () => {
  const container = document.createElement("div");
  container.id = "root";
  document.body.append(container);
  await act(async () => {
    await import("./main.ts");
  });
  expect(container.innerHTML).toBe("<main></main>");
});
