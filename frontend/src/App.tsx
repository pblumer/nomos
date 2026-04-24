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

  useEffect(() => {
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
    <main>
      <h1>Nomos MVP</h1>
      <p>Produktkatalog</p>
      {toast ? (
        <div role="status" aria-label={toast.type === "success" ? "toast-success" : "toast-error"}>
          {toast.message}
        </div>
      ) : null}
      <ul>
        {products.map((product) => (
          <li key={product.id}>
            <button type="button" onClick={() => void loadProductDetail(product.id)}>
              {product.name}
            </button>
          </li>
        ))}
      </ul>

      {selectedProduct ? (
        <section>
          <h2>Produktdetail</h2>
          <p>{selectedProduct.name}</p>
          <p>Version: {selectedProduct.version ?? "n/a"}</p>
          <p>{selectedProduct.description}</p>

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
