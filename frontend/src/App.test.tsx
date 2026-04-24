import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "./App";

afterEach(() => {
  vi.restoreAllMocks();
  cleanup();
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

    fireEvent.click(await screen.findByRole("button", { name: "Rules tab" }));
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

  it("loads requirements and supports filtering", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);

      if (url.includes("/api/v1/products/produkt-1/requirements")) {
        return {
          ok: true,
          json: async () => ({
            product_id: "produkt-1",
            items: ["mailbox-enabled", "login-required", "account-owner-present"],
            count: 3,
          }),
        } as Response;
      }

      if (url.includes("/api/v1/products/produkt-1")) {
        return {
          ok: true,
          json: async () => ({
            id: "produkt-1",
            name: "Benutzerkonto mit Mailbox",
            version: "0.1.0",
            description: "Reference product for Nomos MVP",
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
    fireEvent.click(await screen.findByRole("button", { name: "Load requirements" }));

    expect(await screen.findByText("account-owner-present")).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText("Filter requirements"), {
      target: { value: "login" },
    });

    expect(screen.getByText("login-required")).toBeInTheDocument();
    expect(screen.queryByText("mailbox-enabled")).not.toBeInTheDocument();
  });

  it("supports rules tab with filtering and sorting", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);

      if (url.includes("/api/v1/products/produkt-1/rules")) {
        return {
          ok: true,
          json: async () => ({
            product_id: "produkt-1",
            items: ["rule-zeta", "rule-alpha", "rule-gamma"],
            count: 3,
          }),
        } as Response;
      }

      if (url.includes("/api/v1/products/produkt-1")) {
        return {
          ok: true,
          json: async () => ({
            id: "produkt-1",
            name: "Benutzerkonto mit Mailbox",
            version: "0.1.0",
            description: "Reference product for Nomos MVP",
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
    fireEvent.click(await screen.findByRole("button", { name: "Rules tab" }));
    fireEvent.click(await screen.findByRole("button", { name: "Load rules" }));

    expect(await screen.findByText("rule-zeta")).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText("Filter rules"), {
      target: { value: "alpha" },
    });

    expect(screen.getByText("rule-alpha")).toBeInTheDocument();
    expect(screen.queryByText("rule-zeta")).not.toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("Rules sort"), {
      target: { value: "asc" },
    });

    const sortedItems = screen.getAllByTestId("rules-item").map((node) => node.textContent ?? "");
    expect(sortedItems[0]).toContain("rule-alpha");
  });

  it("shows summary cards in product detail header", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);

      if (url.includes("/api/v1/products/produkt-1/summary")) {
        return {
          ok: true,
          json: async () => ({
            product_id: "produkt-1",
            name: "Benutzerkonto mit Mailbox",
            version: "0.1.0",
            requirements_count: 2,
            rules_count: 3,
            validation_is_valid: true,
            validation_error_count: 0,
          }),
        } as Response;
      }

      if (url.includes("/api/v1/products/produkt-1")) {
        return {
          ok: true,
          json: async () => ({
            id: "produkt-1",
            name: "Benutzerkonto mit Mailbox",
            version: "0.1.0",
            description: "Reference product for Nomos MVP",
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

    expect(await screen.findByText("Requirements: 2")).toBeInTheDocument();
    expect(await screen.findByText("Rules: 3")).toBeInTheDocument();
    expect(await screen.findByText("Validation errors: 0")).toBeInTheDocument();
  });

  it("allows adding and removing requirements and rules", async () => {
    const requirements = ["mailbox-enabled", "login-required"];
    const rules = ["rule-password-policy", "rule-mailbox-quotas"];

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const url = String(input);
      const method = init?.method ?? "GET";

      if (url.includes("/api/v1/products/produkt-1/summary")) {
        return {
          ok: true,
          json: async () => ({
            product_id: "produkt-1",
            name: "Benutzerkonto mit Mailbox",
            version: "0.1.0",
            requirements_count: requirements.length,
            rules_count: rules.length,
            validation_is_valid: true,
            validation_error_count: 0,
          }),
        } as Response;
      }

      if (url.includes("/api/v1/products/produkt-1/requirements") && method === "GET") {
        return {
          ok: true,
          json: async () => ({ product_id: "produkt-1", items: [...requirements], count: requirements.length }),
        } as Response;
      }

      if (url.includes("/api/v1/products/produkt-1/requirements") && method === "POST") {
        requirements.push("mfa-required");
        return { ok: true, json: async () => ({ item: "mfa-required" }) } as Response;
      }

      if (url.includes("/api/v1/products/produkt-1/requirements/mfa-required") && method === "DELETE") {
        const index = requirements.indexOf("mfa-required");
        if (index >= 0) requirements.splice(index, 1);
        return { ok: true, json: async () => ({ item: "mfa-required", removed: true }) } as Response;
      }

      if (url.includes("/api/v1/products/produkt-1/rules") && method === "GET") {
        return {
          ok: true,
          json: async () => ({ product_id: "produkt-1", items: [...rules], count: rules.length }),
        } as Response;
      }

      if (url.includes("/api/v1/products/produkt-1/rules") && method === "POST") {
        rules.push("rule-mfa-enforcement");
        return { ok: true, json: async () => ({ item: "rule-mfa-enforcement" }) } as Response;
      }

      if (url.includes("/api/v1/products/produkt-1/rules/rule-mfa-enforcement") && method === "DELETE") {
        const index = rules.indexOf("rule-mfa-enforcement");
        if (index >= 0) rules.splice(index, 1);
        return { ok: true, json: async () => ({ item: "rule-mfa-enforcement", removed: true }) } as Response;
      }

      if (url.includes("/api/v1/products/produkt-1")) {
        return {
          ok: true,
          json: async () => ({
            id: "produkt-1",
            name: "Benutzerkonto mit Mailbox",
            version: "0.1.0",
            description: "Reference product for Nomos MVP",
            validation: { is_valid: true, errors: [] },
          }),
        } as Response;
      }

      return {
        ok: true,
        json: async () => ({
          items: [
            { id: "produkt-1", name: "Benutzerkonto mit Mailbox", version: "0.1.0" },
          ],
          count: 1,
        }),
      } as Response;
    });

    render(<App />);

    fireEvent.click(await screen.findByRole("button", { name: "Benutzerkonto mit Mailbox" }));
    fireEvent.click(await screen.findByRole("button", { name: "Load requirements" }));

    fireEvent.change(screen.getByPlaceholderText("New requirement"), { target: { value: "mfa-required" } });
    fireEvent.click(screen.getByRole("button", { name: "Add requirement" }));
    expect(await screen.findByText("mfa-required")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Remove requirement mfa-required" }));
    await waitFor(() => expect(screen.queryByText("mfa-required")).not.toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "Rules tab" }));
    fireEvent.click(screen.getByRole("button", { name: "Load rules" }));

    fireEvent.change(screen.getByPlaceholderText("New rule"), { target: { value: "rule-mfa-enforcement" } });
    fireEvent.click(screen.getByRole("button", { name: "Add rule" }));
    expect(await screen.findByText("rule-mfa-enforcement")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Remove rule rule-mfa-enforcement" }));
    await waitFor(() => expect(screen.queryByText("rule-mfa-enforcement")).not.toBeInTheDocument());
  });

  it("shows success and error toasts for requirement actions", async () => {
    const requirements = ["mailbox-enabled", "login-required"];

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const url = String(input);
      const method = init?.method ?? "GET";

      if (url.includes("/api/v1/products/produkt-1/summary")) {
        return {
          ok: true,
          json: async () => ({
            product_id: "produkt-1",
            name: "Benutzerkonto mit Mailbox",
            version: "0.1.0",
            requirements_count: requirements.length,
            rules_count: 2,
            validation_is_valid: true,
            validation_error_count: 0,
          }),
        } as Response;
      }

      if (url.includes("/api/v1/products/produkt-1/requirements") && method === "GET") {
        return {
          ok: true,
          json: async () => ({ product_id: "produkt-1", items: [...requirements], count: requirements.length }),
        } as Response;
      }

      if (url.includes("/api/v1/products/produkt-1/requirements") && method === "POST") {
        const body = JSON.parse(String(init?.body ?? "{}")) as { value?: string };
        if (body.value === "login-required") {
          return {
            ok: false,
            status: 409,
            json: async () => ({ detail: "item already exists" }),
          } as Response;
        }

        requirements.push(String(body.value));
        return { ok: true, json: async () => ({ item: body.value }) } as Response;
      }

      if (url.endsWith("/api/v1/products")) {
        return {
          ok: true,
          json: async () => ({
            items: [{ id: "produkt-1", name: "Benutzerkonto mit Mailbox", version: "0.1.0" }],
            count: 1,
          }),
        } as Response;
      }

      if (url.includes("/api/v1/products/produkt-1")) {
        return {
          ok: true,
          json: async () => ({
            id: "produkt-1",
            name: "Benutzerkonto mit Mailbox",
            version: "0.1.0",
            description: "Reference product for Nomos MVP",
            validation: { is_valid: true, errors: [] },
          }),
        } as Response;
      }

      return { ok: true, json: async () => ({ items: [], count: 0 }) } as Response;
    });

    render(<App />);

    fireEvent.click(await screen.findByRole("button", { name: "Benutzerkonto mit Mailbox" }));
    fireEvent.click(await screen.findByRole("button", { name: "Load requirements" }));

    fireEvent.change(screen.getByPlaceholderText("New requirement"), { target: { value: "mfa-required" } });
    fireEvent.click(screen.getByRole("button", { name: "Add requirement" }));
    expect(await screen.findByRole("status", { name: "toast-success" })).toHaveTextContent("Requirement added");

    fireEvent.change(screen.getByPlaceholderText("New requirement"), { target: { value: "login-required" } });
    fireEvent.click(screen.getByRole("button", { name: "Add requirement" }));
    expect(await screen.findByRole("status", { name: "toast-error" })).toHaveTextContent("item already exists");
  });
});
