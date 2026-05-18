import React from "react";
import ReactDOM from "react-dom/client";

import App from "./App";
import "./index.css";

const TOKEN_STORAGE_KEY = "nomos.auth.token";
const originalFetch = window.fetch.bind(window);

window.fetch = async (input, init) => {
  const url = typeof input === "string" ? input : input instanceof URL ? input.toString() : input.url;
  const isSameOrigin = url.startsWith("/") || url.startsWith(window.location.origin);
  const headers = new Headers(init?.headers ?? (input instanceof Request ? input.headers : undefined));
  if (isSameOrigin) {
    const token = localStorage.getItem(TOKEN_STORAGE_KEY);
    if (token && !headers.has("Authorization")) {
      headers.set("Authorization", `Bearer ${token}`);
    }
  }
  const response = await originalFetch(input, { ...init, headers });
  if (response.status === 401 && isSameOrigin && !url.includes("/api/v1/auth/")) {
    localStorage.removeItem(TOKEN_STORAGE_KEY);
    window.dispatchEvent(new CustomEvent("nomos:unauthorized"));
  }
  return response;
};

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
