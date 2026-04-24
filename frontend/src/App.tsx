import { useEffect, useMemo, useState } from "react";

type Product = {
  id: string;
  name: string;
  version: string;
};

type ProductResponse = {
  items: Product[];
  count: number;
};

type ProductValidation = {
  is_valid: boolean;
  errors: string[];
};

type ProductDetail = {
  id: string;
  name: string;
  version?: string;
  description?: string;
  requirements?: string[];
  rules?: string[];
  validation?: ProductValidation;
};

type ProductRequirementsResponse = {
  product_id: string;
  items: string[];
  count: number;
};

type ProductRulesResponse = {
  product_id: string;
  items: string[];
  count: number;
};

type ProductSummary = {
  product_id: string;
  name: string;
  version: string;
  requirements_count: number;
  rules_count: number;
  validation_is_valid: boolean;
  validation_error_count: number;
};

function App() {
  const [products, setProducts] = useState<Product[]>([]);
  const [selectedProduct, setSelectedProduct] = useState<ProductDetail | null>(null);
  const [productSummary, setProductSummary] = useState<ProductSummary | null>(null);

  const [activeTab, setActiveTab] = useState<"requirements" | "rules">("requirements");

  const [requirements, setRequirements] = useState<string[]>([]);
  const [requirementsFilter, setRequirementsFilter] = useState("");
  const [newRequirement, setNewRequirement] = useState("");

  const [rules, setRules] = useState<string[]>([]);
  const [rulesFilter, setRulesFilter] = useState("");
  const [rulesSort, setRulesSort] = useState<"asc" | "desc">("asc");
  const [newRule, setNewRule] = useState("");

  const [toast, setToast] = useState<{ type: "success" | "error"; message: string } | null>(null);

  const [newProduct, setNewProduct] = useState({
    id: "",
    name: "",
    version: "",
    description: "",
  });
  const [editProduct, setEditProduct] = useState({
    name: "",
    version: "",
    description: "",
  });

  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "";

  const readErrorMessage = async (response: Response) => {
    try {
      const payload = (await response.json()) as { detail?: string };
      if (payload?.detail) return payload.detail;
    } catch {
      // ignore parsing issues
    }
    return "Request failed";
  };

  const loadProducts = async () => {
    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products`);
      if (!response.ok) return;

      const data = (await response.json()) as ProductResponse;
      setProducts(data.items);
    } catch {
      // In dev without backend the UI should still render.
    }
  };

  useEffect(() => {
    void loadProducts();
  }, [apiBaseUrl]);

  const loadProductSummary = async (productId: string) => {
    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products/${productId}/summary`);
      if (!response.ok) return;

      const data = (await response.json()) as ProductSummary;
      setProductSummary(data);
    } catch {
      // Keep UI responsive if summary call fails.
    }
  };

  const loadProductDetail = async (productId: string) => {
    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products/${productId}`);
      if (!response.ok) return;

      const data = (await response.json()) as ProductDetail;
      setSelectedProduct(data);
      setEditProduct({
        name: data.name ?? "",
        version: data.version ?? "",
        description: data.description ?? "",
      });
      setProductSummary(null);
      setActiveTab("requirements");
      setRequirements([]);
      setRequirementsFilter("");
      setNewRequirement("");
      setRules([]);
      setRulesFilter("");
      setRulesSort("asc");
      setNewRule("");
      void loadProductSummary(productId);
    } catch {
      // Keep UI responsive if detail call fails.
    }
  };

  const createProduct = async () => {
    const payload = {
      id: newProduct.id.trim(),
      name: newProduct.name.trim(),
      version: newProduct.version.trim(),
      description: newProduct.description.trim(),
    };

    if (!payload.id || !payload.name || !payload.version) {
      setToast({ type: "error", message: "id, name and version are required" });
      return;
    }

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        setToast({ type: "error", message: await readErrorMessage(response) });
        return;
      }

      setNewProduct({ id: "", name: "", version: "", description: "" });
      await loadProducts();
      await loadProductDetail(payload.id);
      setToast({ type: "success", message: "Product created" });
    } catch {
      setToast({ type: "error", message: "Request failed" });
    }
  };

  const saveProduct = async () => {
    if (!selectedProduct) return;

    const payload = {
      name: editProduct.name.trim(),
      version: editProduct.version.trim(),
      description: editProduct.description.trim(),
    };

    if (!payload.name || !payload.version) {
      setToast({ type: "error", message: "name and version are required" });
      return;
    }

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        setToast({ type: "error", message: await readErrorMessage(response) });
        return;
      }

      await loadProducts();
      await loadProductDetail(selectedProduct.id);
      setToast({ type: "success", message: "Product saved" });
    } catch {
      setToast({ type: "error", message: "Request failed" });
    }
  };

  const removeProduct = async () => {
    if (!selectedProduct) return;

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}`, {
        method: "DELETE",
      });
      if (!response.ok) {
        setToast({ type: "error", message: await readErrorMessage(response) });
        return;
      }

      setSelectedProduct(null);
      setProductSummary(null);
      setRequirements([]);
      setRules([]);
      await loadProducts();
      setToast({ type: "success", message: "Product deleted" });
    } catch {
      setToast({ type: "error", message: "Request failed" });
    }
  };

  const loadRequirements = async () => {
    if (!selectedProduct) return;

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/requirements`);
      if (!response.ok) return;

      const data = (await response.json()) as ProductRequirementsResponse;
      setRequirements(data.items);
    } catch {
      // Keep UI responsive if requirements call fails.
    }
  };

  const addRequirement = async () => {
    if (!selectedProduct) return;
    const value = newRequirement.trim();
    if (value === "") return;

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/requirements`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ value }),
      });
      if (!response.ok) {
        setToast({ type: "error", message: await readErrorMessage(response) });
        return;
      }

      setNewRequirement("");
      await loadRequirements();
      await loadProductSummary(selectedProduct.id);
      setToast({ type: "success", message: "Requirement added" });
    } catch {
      setToast({ type: "error", message: "Request failed" });
    }
  };

  const removeRequirement = async (item: string) => {
    if (!selectedProduct) return;

    try {
      const response = await fetch(
        `${apiBaseUrl}/api/v1/products/${selectedProduct.id}/requirements/${encodeURIComponent(item)}`,
        { method: "DELETE" }
      );
      if (!response.ok) {
        setToast({ type: "error", message: await readErrorMessage(response) });
        return;
      }

      await loadRequirements();
      await loadProductSummary(selectedProduct.id);
      setToast({ type: "success", message: "Requirement removed" });
    } catch {
      setToast({ type: "error", message: "Request failed" });
    }
  };

  const loadRules = async () => {
    if (!selectedProduct) return;

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/rules`);
      if (!response.ok) return;

      const data = (await response.json()) as ProductRulesResponse;
      setRules(data.items);
    } catch {
      // Keep UI responsive if rules call fails.
    }
  };

  const addRule = async () => {
    if (!selectedProduct) return;
    const value = newRule.trim();
    if (value === "") return;

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/rules`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ value }),
      });
      if (!response.ok) {
        setToast({ type: "error", message: await readErrorMessage(response) });
        return;
      }

      setNewRule("");
      await loadRules();
      await loadProductSummary(selectedProduct.id);
      setToast({ type: "success", message: "Rule added" });
    } catch {
      setToast({ type: "error", message: "Request failed" });
    }
  };

  const removeRule = async (item: string) => {
    if (!selectedProduct) return;

    try {
      const response = await fetch(
        `${apiBaseUrl}/api/v1/products/${selectedProduct.id}/rules/${encodeURIComponent(item)}`,
        { method: "DELETE" }
      );
      if (!response.ok) {
        setToast({ type: "error", message: await readErrorMessage(response) });
        return;
      }

      await loadRules();
      await loadProductSummary(selectedProduct.id);
      setToast({ type: "success", message: "Rule removed" });
    } catch {
      setToast({ type: "error", message: "Request failed" });
    }
  };

  const visibleRequirements = useMemo(() => {
    const source = requirements.length > 0 ? requirements : selectedProduct?.requirements ?? [];
    if (requirementsFilter.trim() === "") return source;

    const query = requirementsFilter.toLowerCase();
    return source.filter((item) => item.toLowerCase().includes(query));
  }, [requirements, requirementsFilter, selectedProduct]);

  const visibleRules = useMemo(() => {
    const source = rules.length > 0 ? rules : selectedProduct?.rules ?? [];
    const filtered = rulesFilter.trim()
      ? source.filter((item) => item.toLowerCase().includes(rulesFilter.toLowerCase()))
      : source;

    const sorted = [...filtered].sort((left, right) => left.localeCompare(right));
    return rulesSort === "asc" ? sorted : sorted.reverse();
  }, [rules, rulesFilter, rulesSort, selectedProduct]);

  return (
    <main className="app-shell">
      <section className="card stack">
        <h1>Nomos MVP</h1>
        <p className="muted">Produktkatalog</p>
      </section>
      {toast ? (
        <div
          className={`card ${toast.type === "success" ? "toast-success" : "toast-error"}`}
          role="status"
          aria-label={toast.type === "success" ? "toast-success" : "toast-error"}
        >
          {toast.message}
        </div>
      ) : null}

      <section className="card stack">
        <h2>Neues Produkt</h2>
        <input
          type="text"
          placeholder="New product id"
          value={newProduct.id}
          onChange={(event) => setNewProduct((current) => ({ ...current, id: event.target.value }))}
        />
        <input
          type="text"
          placeholder="New product name"
          value={newProduct.name}
          onChange={(event) => setNewProduct((current) => ({ ...current, name: event.target.value }))}
        />
        <input
          type="text"
          placeholder="New product version"
          value={newProduct.version}
          onChange={(event) => setNewProduct((current) => ({ ...current, version: event.target.value }))}
        />
        <input
          type="text"
          placeholder="New product description"
          value={newProduct.description}
          onChange={(event) => setNewProduct((current) => ({ ...current, description: event.target.value }))}
        />
        <button type="button" onClick={() => void createProduct()}>
          Create product
        </button>
      </section>

      <section className="card stack">
        <h2>Produkte</h2>
        <ul className="list">
          {products.map((product) => (
            <li key={product.id} className="list-item">
              <button type="button" onClick={() => void loadProductDetail(product.id)}>
                {product.name}
              </button>
              <span className="muted">{product.version}</span>
            </li>
          ))}
        </ul>
      </section>

      {selectedProduct ? (
        <section className="card stack">
          <h2>Produktdetail</h2>
          <p>{selectedProduct.name}</p>
          <p>Version: {selectedProduct.version ?? "n/a"}</p>
          <p>{selectedProduct.description}</p>

          <section>
            <h3>Produkt bearbeiten</h3>
            <input
              type="text"
              placeholder="Edit product name"
              value={editProduct.name}
              onChange={(event) => setEditProduct((current) => ({ ...current, name: event.target.value }))}
            />
            <input
              type="text"
              placeholder="Edit product version"
              value={editProduct.version}
              onChange={(event) => setEditProduct((current) => ({ ...current, version: event.target.value }))}
            />
            <input
              type="text"
              placeholder="Edit product description"
              value={editProduct.description}
              onChange={(event) => setEditProduct((current) => ({ ...current, description: event.target.value }))}
            />
            <button type="button" onClick={() => void saveProduct()}>
              Save product
            </button>
            <button type="button" onClick={() => void removeProduct()}>
              Delete product
            </button>
          </section>

          {productSummary ? (
            <div>
              <p>Requirements: {productSummary.requirements_count}</p>
              <p>Rules: {productSummary.rules_count}</p>
              <p>Validation errors: {productSummary.validation_error_count}</p>
            </div>
          ) : null}

          <div>
            <button type="button" onClick={() => setActiveTab("requirements")}>
              Requirements tab
            </button>
            <button type="button" onClick={() => setActiveTab("rules")}>
              Rules tab
            </button>
          </div>

          {activeTab === "requirements" ? (
            <section>
              <h3>Requirements</h3>
              <button type="button" onClick={() => void loadRequirements()}>
                Load requirements
              </button>
              <div>
                <input
                  type="text"
                  placeholder="Filter requirements"
                  value={requirementsFilter}
                  onChange={(event) => setRequirementsFilter(event.target.value)}
                />
              </div>
              <div>
                <input
                  type="text"
                  placeholder="New requirement"
                  value={newRequirement}
                  onChange={(event) => setNewRequirement(event.target.value)}
                />
                <button type="button" onClick={() => void addRequirement()}>
                  Add requirement
                </button>
              </div>
              <ul>
                {visibleRequirements.map((requirement) => (
                  <li key={requirement}>
                    <span>{requirement}</span>
                    <button type="button" onClick={() => void removeRequirement(requirement)}>
                      Remove requirement {requirement}
                    </button>
                  </li>
                ))}
              </ul>
            </section>
          ) : (
            <section>
              <h3>Rules</h3>
              <button type="button" onClick={() => void loadRules()}>
                Load rules
              </button>
              <div>
                <input
                  type="text"
                  placeholder="Filter rules"
                  value={rulesFilter}
                  onChange={(event) => setRulesFilter(event.target.value)}
                />
              </div>
              <div>
                <input
                  type="text"
                  placeholder="New rule"
                  value={newRule}
                  onChange={(event) => setNewRule(event.target.value)}
                />
                <button type="button" onClick={() => void addRule()}>
                  Add rule
                </button>
              </div>
              <div>
                <label htmlFor="rules-sort">Rules sort</label>
                <select
                  id="rules-sort"
                  value={rulesSort}
                  onChange={(event) => setRulesSort(event.target.value as "asc" | "desc")}
                >
                  <option value="asc">asc</option>
                  <option value="desc">desc</option>
                </select>
              </div>
              <ul>
                {visibleRules.map((rule) => (
                  <li key={rule} data-testid="rules-item">
                    <span>{rule}</span>
                    <button type="button" onClick={() => void removeRule(rule)}>
                      Remove rule {rule}
                    </button>
                  </li>
                ))}
              </ul>
            </section>
          )}

          <h3>Validation</h3>
          {selectedProduct.validation?.errors?.length ? (
            <ul>
              {selectedProduct.validation.errors.map((errorMessage) => (
                <li key={errorMessage}>{errorMessage}</li>
              ))}
            </ul>
          ) : (
            <p>No validation errors</p>
          )}
        </section>
      ) : null}
    </main>
  );
}

export default App;
