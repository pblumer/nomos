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

type RuleDetail = {
  id: string;
  name: string;
  description?: string;
  category?: string;
  severity?: string;
  scope?: string;
  enforcement?: string;
  validation?: {
    type?: string;
    field?: string;
    operator?: string;
    value?: string | number | boolean;
  };
};

type RuleEditorState = {
  name: string;
  description: string;
  category: string;
  severity: string;
  scope: string;
  enforcement: string;
  validationType: string;
  validationField: string;
  validationOperator: string;
  validationValue: string;
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
  const [selectedRuleId, setSelectedRuleId] = useState<string | null>(null);
  const [ruleDetails, setRuleDetails] = useState<Record<string, RuleDetail>>({});
  const [ruleEditor, setRuleEditor] = useState<RuleEditorState>({
    name: "",
    description: "",
    category: "",
    severity: "",
    scope: "",
    enforcement: "",
    validationType: "",
    validationField: "",
    validationOperator: "",
    validationValue: "",
  });

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

  const toRuleEditorState = (detail: RuleDetail): RuleEditorState => ({
    name: detail.name ?? "",
    description: detail.description ?? "",
    category: detail.category ?? "",
    severity: detail.severity ?? "",
    scope: detail.scope ?? "",
    enforcement: detail.enforcement ?? "",
    validationType: detail.validation?.type ? String(detail.validation.type) : "",
    validationField: detail.validation?.field ? String(detail.validation.field) : "",
    validationOperator: detail.validation?.operator ? String(detail.validation.operator) : "",
    validationValue: detail.validation?.value !== undefined ? String(detail.validation.value) : "",
  });

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
      setSelectedRuleId(null);
      setRuleDetails({});
      setRuleEditor({
        name: "",
        description: "",
        category: "",
        severity: "",
        scope: "",
        enforcement: "",
        validationType: "",
        validationField: "",
        validationOperator: "",
        validationValue: "",
      });
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

    let confirmed = true;
    try {
      confirmed = typeof window.confirm === "function" ? window.confirm(`Delete product ${selectedProduct.name}?`) : true;
    } catch {
      confirmed = true;
    }
    if (!confirmed) return;

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
      setSelectedRuleId(null);
      setRuleDetails({});
      setRuleEditor({
        name: "",
        description: "",
        category: "",
        severity: "",
        scope: "",
        enforcement: "",
        validationType: "",
        validationField: "",
        validationOperator: "",
        validationValue: "",
      });
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
      setSelectedRuleId(null);
      setRuleEditor({
        name: "",
        description: "",
        category: "",
        severity: "",
        scope: "",
        enforcement: "",
        validationType: "",
        validationField: "",
        validationOperator: "",
        validationValue: "",
      });
    } catch {
      // Keep UI responsive if rules call fails.
    }
  };

  const loadRuleDetail = async (ruleId: string) => {
    if (ruleDetails[ruleId]) {
      setSelectedRuleId(ruleId);
      setRuleEditor(toRuleEditorState(ruleDetails[ruleId]));
      return;
    }

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/rules/${ruleId}`);
      if (!response.ok) {
        setToast({ type: "error", message: await readErrorMessage(response) });
        return;
      }

      const data = (await response.json()) as RuleDetail;
      setRuleDetails((current) => ({ ...current, [ruleId]: data }));
      setSelectedRuleId(ruleId);
      setRuleEditor(toRuleEditorState(data));
    } catch {
      setToast({ type: "error", message: "Request failed" });
    }
  };

  const saveRuleDetail = async () => {
    if (!selectedRuleId) return;

    const payload: RuleDetail = {
      id: selectedRuleId,
      name: ruleEditor.name.trim(),
      description: ruleEditor.description.trim(),
      category: ruleEditor.category.trim(),
      severity: ruleEditor.severity.trim(),
      scope: ruleEditor.scope.trim(),
      enforcement: ruleEditor.enforcement.trim(),
    };

    if (ruleEditor.validationType.trim() || ruleEditor.validationField.trim() || ruleEditor.validationOperator.trim() || ruleEditor.validationValue.trim()) {
      payload.validation = {
        type: ruleEditor.validationType.trim(),
        field: ruleEditor.validationField.trim(),
        operator: ruleEditor.validationOperator.trim(),
        value: ruleEditor.validationValue.trim(),
      };
    }

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/rules/${selectedRuleId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        setToast({ type: "error", message: await readErrorMessage(response) });
        return;
      }

      const data = (await response.json()) as RuleDetail;
      setRuleDetails((current) => ({ ...current, [selectedRuleId]: data }));
      setRuleEditor(toRuleEditorState(data));
      setToast({ type: "success", message: "Rule saved" });
    } catch {
      setToast({ type: "error", message: "Request failed" });
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

  const selectedRuleDetail = selectedRuleId ? ruleDetails[selectedRuleId] ?? null : null;

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

      <section className="workspace-grid">
        <div className="left-column">
          <section className="card stack">
            <h2 className="section-title">Neues Produkt</h2>
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
            <h2 className="section-title">Produkte</h2>
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
        </div>

        <div className="right-column">
          {selectedProduct ? (
            <section className="card stack">
              <h2 className="section-title">Produktdetail</h2>
              <p>{selectedProduct.name}</p>
              <p>Version: {selectedProduct.version ?? "n/a"}</p>
              <p>{selectedProduct.description}</p>

              <section className="stack">
                <h3 className="section-title">Produkt bearbeiten</h3>
                <div className="row">
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
                </div>
                <div className="list-item-actions">
                  <button type="button" onClick={() => void saveProduct()}>
                    Save product
                  </button>
                  <button className="secondary" type="button" onClick={() => void removeProduct()}>
                    Delete product
                  </button>
                </div>
              </section>

              {productSummary ? (
                <div className="stats-grid">
                  <div className="stat">Requirements: {productSummary.requirements_count}</div>
                  <div className="stat">Rules: {productSummary.rules_count}</div>
                  <div className="stat">Validation errors: {productSummary.validation_error_count}</div>
                </div>
              ) : null}

              <div className="list-item-actions">
                <button type="button" onClick={() => setActiveTab("requirements")}>
                  Requirements tab
                </button>
                <button type="button" onClick={() => setActiveTab("rules")}>
                  Rules tab
                </button>
              </div>

              {activeTab === "requirements" ? (
                <section className="stack">
                  <h3 className="section-title">Requirements</h3>
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
                  <div className="list-item-actions">
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
                  <ul className="list">
                    {visibleRequirements.map((requirement) => (
                      <li key={requirement} className="list-item">
                        <span>{requirement}</span>
                        <button className="secondary" type="button" onClick={() => void removeRequirement(requirement)}>
                          Remove requirement {requirement}
                        </button>
                      </li>
                    ))}
                  </ul>
                </section>
              ) : (
                <section className="stack">
                  <h3 className="section-title">Rules</h3>
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
                  <div className="list-item-actions">
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
                  <ul className="list">
                    {visibleRules.map((rule) => (
                      <li key={rule} className="list-item" data-testid="rules-item">
                        <span>{rule}</span>
                        <div className="list-item-actions">
                          <button className="secondary" type="button" onClick={() => void loadRuleDetail(rule)}>
                            Show rule details {rule}
                          </button>
                          <button className="secondary" type="button" onClick={() => void removeRule(rule)}>
                            Remove rule {rule}
                          </button>
                        </div>
                      </li>
                    ))}
                  </ul>

                  {selectedRuleDetail ? (
                    <section className="card stack">
                      <h4 className="section-title">Rule detail: {selectedRuleDetail.name}</h4>
                      <p>ID: {selectedRuleDetail.id}</p>

                      <input
                        type="text"
                        placeholder="Edit rule name"
                        value={ruleEditor.name}
                        onChange={(event) => setRuleEditor((current) => ({ ...current, name: event.target.value }))}
                      />
                      <input
                        type="text"
                        placeholder="Edit rule description"
                        value={ruleEditor.description}
                        onChange={(event) => setRuleEditor((current) => ({ ...current, description: event.target.value }))}
                      />
                      <div className="row">
                        <input
                          type="text"
                          placeholder="Edit rule category"
                          value={ruleEditor.category}
                          onChange={(event) => setRuleEditor((current) => ({ ...current, category: event.target.value }))}
                        />
                        <input
                          type="text"
                          placeholder="Edit rule severity"
                          value={ruleEditor.severity}
                          onChange={(event) => setRuleEditor((current) => ({ ...current, severity: event.target.value }))}
                        />
                        <input
                          type="text"
                          placeholder="Edit rule scope"
                          value={ruleEditor.scope}
                          onChange={(event) => setRuleEditor((current) => ({ ...current, scope: event.target.value }))}
                        />
                      </div>
                      <input
                        type="text"
                        placeholder="Edit rule enforcement"
                        value={ruleEditor.enforcement}
                        onChange={(event) => setRuleEditor((current) => ({ ...current, enforcement: event.target.value }))}
                      />

                      <div className="row">
                        <input
                          type="text"
                          placeholder="Edit validation type"
                          value={ruleEditor.validationType}
                          onChange={(event) => setRuleEditor((current) => ({ ...current, validationType: event.target.value }))}
                        />
                        <input
                          type="text"
                          placeholder="Edit validation field"
                          value={ruleEditor.validationField}
                          onChange={(event) => setRuleEditor((current) => ({ ...current, validationField: event.target.value }))}
                        />
                        <input
                          type="text"
                          placeholder="Edit validation operator"
                          value={ruleEditor.validationOperator}
                          onChange={(event) => setRuleEditor((current) => ({ ...current, validationOperator: event.target.value }))}
                        />
                        <input
                          type="text"
                          placeholder="Edit validation value"
                          value={ruleEditor.validationValue}
                          onChange={(event) => setRuleEditor((current) => ({ ...current, validationValue: event.target.value }))}
                        />
                      </div>

                      <button type="button" onClick={() => void saveRuleDetail()}>
                        Save rule
                      </button>
                    </section>
                  ) : null}
                </section>
              )}

              <h3 className="section-title">Validation</h3>
              {selectedProduct.validation?.errors?.length ? (
                <ul className="list">
                  {selectedProduct.validation.errors.map((errorMessage) => (
                    <li className="list-item" key={errorMessage}>
                      {errorMessage}
                    </li>
                  ))}
                </ul>
              ) : (
                <p>No validation errors</p>
              )}
            </section>
          ) : (
            <section className="card">
              <p className="muted">Bitte links ein Produkt wählen.</p>
            </section>
          )}
        </div>
      </section>
    </main>
  );
}

export default App;
