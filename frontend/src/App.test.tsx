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
            description: "Reference product for Nomos MVP",
            requirements: ["mailbox-enabled", "login-required"],
            rules: ["rule-mailbox-quotas", "rule-password-policy"],
            validation: { is_valid: true, errors: [] },
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

    expect(await screen.findByText("Reference product for Nomos MVP")).toBeInTheDocument();
    expect(await screen.findByText("mailbox-enabled")).toBeInTheDocument();
    expect(await screen.findByText("rule-mailbox-quotas")).toBeInTheDocument();
  });

  it("shows validation errors in product detail", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);
      if (url.includes("/api/v1/products/produkt-invalid")) {
        return {
          ok: true,
          json: async () => ({
            id: "produkt-invalid",
            name: "Broken Product",
            description: "Missing version field",
            validation: {
              is_valid: false,
              errors: ["Missing required field: version"],
            },
          }),
        } as Response;
      }

      return {
        ok: true,
        json: async () => ({
          items: [
            {
              id: "produkt-invalid",
              name: "Broken Product",
              version: "",
            },
          ],
          count: 1,
        }),
      } as Response;
    });

    render(<App />);

    fireEvent.click(await screen.findByRole("button", { name: "Broken Product" }));

    expect(await screen.findByText("Missing required field: version")).toBeInTheDocument();
  });
});
