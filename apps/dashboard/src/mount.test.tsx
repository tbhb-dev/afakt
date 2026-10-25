import { test as fcTest } from "@fast-check/vitest";
import fc from "fast-check";
import { act } from "react";
import type { Root } from "react-dom/client";
import { afterEach, expect, test } from "vitest";
import { mount } from "./mount.tsx";

const missingRoot = "the page has no #root element";

function mountError(): unknown {
  let caught: unknown;
  try {
    mount(document);
  } catch (error) {
    caught = error;
  }
  return caught;
}

afterEach(() => {
  document.body.replaceChildren();
});

test("mount renders the blank page into #root", () => {
  const container = document.createElement("div");
  container.id = "root";
  document.body.append(container);
  let root: Root | undefined;
  act(() => {
    root = mount(document);
  });
  expect(container.innerHTML).toBe("<main></main>");
  act(() => {
    root?.unmount();
  });
  expect(container.innerHTML).toBe("");
});

test("mount throws when the page has no elements", () => {
  expect(mountError()).toStrictEqual(new Error(missingRoot));
});

// The property holds the predicate's boolean result to the assertion, so it
// needs no expect() call inside the callback.
fcTest.prop([fc.string().filter((id) => id !== "root")])(
  "mount throws whenever no element has the id root",
  (id) => {
    const element = document.createElement("div");
    element.id = id;
    document.body.replaceChildren(element);
    const error = mountError();
    return error instanceof Error && error.message === missingRoot;
  },
);
