import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "./App";

afterEach(() => {
  vi.restoreAllMocks();
});

describe("App", () => {
  it("shows Nomos title", () => {
    render(<App />);

    expect(screen.getByRole("heading", { name: "Nomos MVP" })).toBeInTheDocument();
  });

  it("loads and renders product names from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => ({
        items: [
          {
            id: "produkt-1",
            name: "Benutzerkonto mit Mailbox",
            version: "0.1.0",
          },
        ],
        count: 1,
      }),
    } as Response);

    render(<App />);

    expect(await screen.findByText("Benutzerkonto mit Mailbox")).toBeInTheDocument();
  });

  it("loads product detail when a product is clicked", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);
      if (url.includes("/api/v1/products/produkt-1")) {
        return {
          ok: true,
          json: async () => ({
            id: "produkt-1",
            name: "Benutzerkonto mit Mailbox",
            version: "0.1.0",
            beschreibung: "Referenzprodukt fuer Nomos MVP",
          }),
        } as Response;
      }

      return {
        ok: true,
        json: async () => ({
          items: [
            {
              id: "produkt-1",
              name: "Benutzerkonto mit Mailbox",
              version: "0.1.0",
            },
          ],
          count: 1,
        }),
      } as Response;
    });

    render(<App />);

    fireEvent.click(await screen.findByRole("button", { name: "Benutzerkonto mit Mailbox" }));

    expect(await screen.findByText("Referenzprodukt fuer Nomos MVP")).toBeInTheDocument();
  });
});
