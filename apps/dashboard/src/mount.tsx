import { StrictMode } from "react";
import { createRoot, type Root } from "react-dom/client";
import { App } from "./app.tsx";

export function mount(document: Document): Root {
  const container = document.querySelector("#root");
  if (container === null) {
    throw new Error("the page has no #root element");
  }
  const root = createRoot(container);
  root.render(
    <StrictMode>
      <App />
    </StrictMode>,
  );
  return root;
}
