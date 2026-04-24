import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import ServerView from "./ServerView";

afterEach(() => {
  vi.restoreAllMocks();
  cleanup();
});

describe("ServerView", () => {
  it("shows server title", () => {
    vi.spyOn(window, "confirm").mockReturnValue(true);
    render(<ServerView apiBaseUrl="" />);
    expect(screen.getByRole("heading", { name: "Server" })).toBeInTheDocument();
  });

  it("loads and renders server names from API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => ({
        items: [{ id: "srv-web-01", name: "Web Server 01" }],
        count: 1,
      }),
    } as Response);

    vi.spyOn(window, "confirm").mockReturnValue(true);
    render(<ServerView apiBaseUrl="" />);

    expect(await screen.findByRole("button", { name: "Web Server 01" })).toBeInTheDocument();
  });

  it("loads server detail when a server is clicked", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = String(input);
      if (url.includes("/api/v1/servers/srv-web-01")) {
        return {
          ok: true,
          json: async () => ({
            id: "srv-web-01",
            name: "Web Server 01",
            description: "Main web server",
            product_ids: ["produkt-1"],
            tags: ["production", "web"],
          }),
        } as Response;
      }

      return {
        ok: true,
        json: async () => ({
          items: [{ id: "srv-web-01", name: "Web Server 01" }],
          count: 1,
        }),
      } as Response;
    });

    vi.spyOn(window, "confirm").mockReturnValue(true);
    render(<ServerView apiBaseUrl="" />);

    fireEvent.click(await screen.findByRole("button", { name: "Web Server 01" }));

    expect(await screen.findByText("Main web server")).toBeInTheDocument();
    expect(await screen.findByText("Products: produkt-1")).toBeInTheDocument();
    expect(await screen.findByText("Tags: production, web")).toBeInTheDocument();
  });

  it("supports creating a server", async () => {
    const servers = [{ id: "srv-a", name: "Server A" }];

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const url = String(input);
      const method = init?.method ?? "GET";

      if (url.endsWith("/api/v1/servers") && method === "GET") {
        return {
          ok: true,
          json: async () => ({ items: servers, count: servers.length }),
        } as Response;
      }

      if (url.endsWith("/api/v1/servers") && method === "POST") {
        const body = JSON.parse(String(init?.body ?? "{}")) as { id: string; name: string };
        servers.push(body);
        return { ok: true, json: async () => body } as Response;
      }

      if (url.includes("/api/v1/servers/srv-b")) {
        return {
          ok: true,
          json: async () => ({ id: "srv-b", name: "Server B" }),
        } as Response;
      }

      return { ok: true, json: async () => ({ items: [], count: 0 }) } as Response;
    });

    vi.spyOn(window, "confirm").mockReturnValue(true);
    render(<ServerView apiBaseUrl="" />);

    fireEvent.change(screen.getByPlaceholderText("New server id"), { target: { value: "srv-b" } });
    fireEvent.change(screen.getByPlaceholderText("New server name"), { target: { value: "Server B" } });
    fireEvent.click(screen.getByRole("button", { name: "Create server" }));

    expect(await screen.findByRole("button", { name: "Server B" })).toBeInTheDocument();
  });

  it("supports editing and deleting a server", async () => {
    const servers = [{ id: "srv-a", name: "Server A", description: "Initial", product_ids: [] as string[], tags: [] as string[] }];

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const url = String(input);
      const method = init?.method ?? "GET";

      if (url.endsWith("/api/v1/servers") && method === "GET") {
        return {
          ok: true,
          json: async () => ({
            items: servers.map((s) => ({ id: s.id, name: s.name })),
            count: servers.length,
          }),
        } as Response;
      }

      if (url.includes("/api/v1/servers/srv-a") && method === "PUT") {
        const body = JSON.parse(String(init?.body ?? "{}")) as { name: string };
        servers[0].name = body.name;
        return { ok: true, json: async () => servers[0] } as Response;
      }

      if (url.includes("/api/v1/servers/srv-a") && method === "DELETE") {
        servers.splice(0, 1);
        return { ok: true, json: async () => ({ id: "srv-a", removed: true }) } as Response;
      }

      if (url.includes("/api/v1/servers/srv-a")) {
        return { ok: true, json: async () => servers[0] } as Response;
      }

      return { ok: true, json: async () => ({ items: [], count: 0 }) } as Response;
    });

    vi.spyOn(window, "confirm").mockReturnValue(true);
    render(<ServerView apiBaseUrl="" />);

    fireEvent.click(await screen.findByRole("button", { name: "Server A" }));
    fireEvent.change(await screen.findByPlaceholderText("Edit server name"), { target: { value: "Server A Updated" } });
    fireEvent.click(screen.getByRole("button", { name: "Save server" }));

    await waitFor(() => expect(screen.getByRole("button", { name: "Server A Updated" })).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "Delete server" }));
    await waitFor(() => expect(screen.queryByRole("button", { name: "Server A Updated" })).not.toBeInTheDocument());
  });

  it("shows frontend validation error when creating a server without required fields", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => ({ items: [], count: 0 }),
    } as Response);

    vi.spyOn(window, "confirm").mockReturnValue(true);
    render(<ServerView apiBaseUrl="" />);

    fireEvent.change(screen.getByPlaceholderText("New server id"), { target: { value: "srv-x" } });
    fireEvent.click(screen.getByRole("button", { name: "Create server" }));

    expect(await screen.findByRole("status", { name: "toast-error" })).toHaveTextContent("id and name are required");
  });
});
