import { renderToStaticMarkup } from "react-dom/server";
import { expect, test } from "vitest";
import { App } from "./app.tsx";

test("App renders an empty main landmark", () => {
  expect(renderToStaticMarkup(<App />)).toBe("<main></main>");
});
