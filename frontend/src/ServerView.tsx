import { useEffect, useState } from "react";

type Server = {
  id: string;
  name: string;
};

type ServerResponse = {
  items: Server[];
  count: number;
};

type ServerDetail = {
  id: string;
  name: string;
  description?: string;
  product_ids?: string[];
  tags?: string[];
};

type ServerEditorState = {
  name: string;
  description: string;
  productIds: string;
  tags: string;
};

export default function ServerView({ apiBaseUrl }: { apiBaseUrl: string }) {
  const [servers, setServers] = useState<Server[]>([]);
  const [selectedServer, setSelectedServer] = useState<ServerDetail | null>(null);
  const [editServer, setEditServer] = useState<ServerEditorState>({
    name: "",
    description: "",
    productIds: "",
    tags: "",
  });
  const [newServer, setNewServer] = useState({
    id: "",
    name: "",
    description: "",
    productIds: "",
    tags: "",
  });
  const [toast, setToast] = useState<{ type: "success" | "error"; message: string } | null>(null);

  const readErrorMessage = async (response: Response) => {
    try {
      const payload = (await response.json()) as { detail?: string };
      if (payload?.detail) return payload.detail;
    } catch {
      // ignore
    }
    return "Request failed";
  };

  const toEditorState = (detail: ServerDetail): ServerEditorState => ({
    name: detail.name ?? "",
    description: detail.description ?? "",
    productIds: (detail.product_ids ?? []).join(", "),
    tags: (detail.tags ?? []).join(", "),
  });

  const loadServers = async () => {
    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/servers`);
      if (!response.ok) return;
      const data = (await response.json()) as ServerResponse;
      setServers(data.items);
    } catch {
      // ignore
    }
  };

  useEffect(() => {
    void loadServers();
  }, [apiBaseUrl]);

  const loadServerDetail = async (serverId: string) => {
    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/servers/${serverId}`);
      if (!response.ok) return;
      const data = (await response.json()) as ServerDetail;
      setSelectedServer(data);
      setEditServer(toEditorState(data));
    } catch {
      // ignore
    }
  };

  const createServer = async () => {
    const payload = {
      id: newServer.id.trim(),
      name: newServer.name.trim(),
      description: newServer.description.trim(),
      product_ids: newServer.productIds.split(",").map((s) => s.trim()).filter(Boolean),
      tags: newServer.tags.split(",").map((s) => s.trim()).filter(Boolean),
    };
    if (!payload.id || !payload.name) {
      setToast({ type: "error", message: "id and name are required" });
      return;
    }
    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/servers`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        setToast({ type: "error", message: await readErrorMessage(response) });
        return;
      }
      setNewServer({ id: "", name: "", description: "", productIds: "", tags: "" });
      await loadServers();
      await loadServerDetail(payload.id);
      setToast({ type: "success", message: "Server created" });
    } catch {
      setToast({ type: "error", message: "Request failed" });
    }
  };

  const saveServer = async () => {
    if (!selectedServer) return;
    const payload = {
      name: editServer.name.trim(),
      description: editServer.description.trim(),
      product_ids: editServer.productIds.split(",").map((s) => s.trim()).filter(Boolean),
      tags: editServer.tags.split(",").map((s) => s.trim()).filter(Boolean),
    };
    if (!payload.name) {
      setToast({ type: "error", message: "name is required" });
      return;
    }
    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/servers/${selectedServer.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        setToast({ type: "error", message: await readErrorMessage(response) });
        return;
      }
      await loadServers();
      await loadServerDetail(selectedServer.id);
      setToast({ type: "success", message: "Server saved" });
    } catch {
      setToast({ type: "error", message: "Request failed" });
    }
  };

  const removeServer = async () => {
    if (!selectedServer) return;
    let confirmed = true;
    try {
      confirmed = typeof window.confirm === "function" ? window.confirm(`Delete server ${selectedServer.name}?`) : true;
    } catch {
      confirmed = true;
    }
    if (!confirmed) return;
    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/servers/${selectedServer.id}`, {
        method: "DELETE",
      });
      if (!response.ok) {
        setToast({ type: "error", message: await readErrorMessage(response) });
        return;
      }
      setSelectedServer(null);
      await loadServers();
      setToast({ type: "success", message: "Server deleted" });
    } catch {
      setToast({ type: "error", message: "Request failed" });
    }
  };

  return (
    <>
      {toast ? (
        <div
          className={`card ${toast.type === "success" ? "toast-success" : "toast-error"}`}
          role="status"
          aria-label={toast.type === "success" ? "toast-success" : "toast-error"}
        >
          {toast.message}
        </div>
      ) : null}
      <section className="workspace-grid">
        <div className="left-column">
          <section className="card stack">
            <h2 className="section-title">Neuer Server</h2>
            <input
              type="text"
              placeholder="New server id"
              value={newServer.id}
              onChange={(event) => setNewServer((current) => ({ ...current, id: event.target.value }))}
            />
            <input
              type="text"
              placeholder="New server name"
              value={newServer.name}
              onChange={(event) => setNewServer((current) => ({ ...current, name: event.target.value }))}
            />
            <input
              type="text"
              placeholder="New server description"
              value={newServer.description}
              onChange={(event) => setNewServer((current) => ({ ...current, description: event.target.value }))}
            />
            <input
              type="text"
              placeholder="Product ids (comma separated)"
              value={newServer.productIds}
              onChange={(event) => setNewServer((current) => ({ ...current, productIds: event.target.value }))}
            />
            <input
              type="text"
              placeholder="Tags (comma separated)"
              value={newServer.tags}
              onChange={(event) => setNewServer((current) => ({ ...current, tags: event.target.value }))}
            />
            <button type="button" onClick={() => void createServer()}>
              Create server
            </button>
          </section>

          <section className="card stack">
            <h2 className="section-title">Server</h2>
            <ul className="list">
              {servers.map((server) => (
                <li key={server.id} className="list-item">
                  <button type="button" onClick={() => void loadServerDetail(server.id)}>
                    {server.name}
                  </button>
                </li>
              ))}
            </ul>
          </section>
        </div>

        <div className="right-column">
          {selectedServer ? (
            <section className="card stack">
              <h2 className="section-title">Serverdetail</h2>
              <p>{selectedServer.name}</p>
              <p>{selectedServer.description}</p>
              {selectedServer.product_ids && selectedServer.product_ids.length > 0 ? (
                <p>Products: {selectedServer.product_ids.join(", ")}</p>
              ) : null}
              {selectedServer.tags && selectedServer.tags.length > 0 ? (
                <p>Tags: {selectedServer.tags.join(", ")}</p>
              ) : null}

              <section className="stack">
                <h3 className="section-title">Server bearbeiten</h3>
                <div className="row">
                  <input
                    type="text"
                    placeholder="Edit server name"
                    value={editServer.name}
                    onChange={(event) => setEditServer((current) => ({ ...current, name: event.target.value }))}
                  />
                </div>
                <input
                  type="text"
                  placeholder="Edit server description"
                  value={editServer.description}
                  onChange={(event) => setEditServer((current) => ({ ...current, description: event.target.value }))}
                />
                <input
                  type="text"
                  placeholder="Product ids (comma separated)"
                  value={editServer.productIds}
                  onChange={(event) => setEditServer((current) => ({ ...current, productIds: event.target.value }))}
                />
                <input
                  type="text"
                  placeholder="Tags (comma separated)"
                  value={editServer.tags}
                  onChange={(event) => setEditServer((current) => ({ ...current, tags: event.target.value }))}
                />
                <div className="list-item-actions">
                  <button type="button" onClick={() => void saveServer()}>
                    Save server
                  </button>
                  <button className="secondary" type="button" onClick={() => void removeServer()}>
                    Delete server
                  </button>
                </div>
              </section>
            </section>
          ) : (
            <section className="card">
              <p className="muted">Bitte links einen Server waehlen.</p>
            </section>
          )}
        </div>
      </section>
    </>
  );
}
