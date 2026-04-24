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

function App() {
  const [products, setProducts] = useState<Product[]>([]);
  const [selectedProduct, setSelectedProduct] = useState<ProductDetail | null>(null);

  const [activeTab, setActiveTab] = useState<"requirements" | "rules">("requirements");

  const [requirements, setRequirements] = useState<string[]>([]);
  const [requirementsFilter, setRequirementsFilter] = useState("");

  const [rules, setRules] = useState<string[]>([]);
  const [rulesFilter, setRulesFilter] = useState("");
  const [rulesSort, setRulesSort] = useState<"asc" | "desc">("asc");

  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "";

  useEffect(() => {
    const loadProducts = async () => {
      try {
        const response = await fetch(`${apiBaseUrl}/api/v1/products`);
        if (!response.ok) {
          return;
        }

        const data = (await response.json()) as ProductResponse;
        setProducts(data.items);
      } catch {
        // In dev without backend the UI should still render.
      }
    };

    void loadProducts();
  }, [apiBaseUrl]);

  const loadProductDetail = async (productId: string) => {
    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products/${productId}`);
      if (!response.ok) {
        return;
      }

      const data = (await response.json()) as ProductDetail;
      setSelectedProduct(data);
      setActiveTab("requirements");
      setRequirements([]);
      setRequirementsFilter("");
      setRules([]);
      setRulesFilter("");
      setRulesSort("asc");
    } catch {
      // Keep UI responsive if detail call fails.
    }
  };

  const loadRequirements = async () => {
    if (!selectedProduct) {
      return;
    }

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/requirements`);
      if (!response.ok) {
        return;
      }

      const data = (await response.json()) as ProductRequirementsResponse;
      setRequirements(data.items);
    } catch {
      // Keep UI responsive if requirements call fails.
    }
  };

  const loadRules = async () => {
    if (!selectedProduct) {
      return;
    }

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/rules`);
      if (!response.ok) {
        return;
      }

      const data = (await response.json()) as ProductRulesResponse;
      setRules(data.items);
    } catch {
      // Keep UI responsive if rules call fails.
    }
  };

  const visibleRequirements = useMemo(() => {
    const source = requirements.length > 0 ? requirements : selectedProduct?.requirements ?? [];

    if (requirementsFilter.trim() === "") {
      return source;
    }

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
              <ul>
                {visibleRequirements.map((requirement) => (
                  <li key={requirement}>{requirement}</li>
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
                    {rule}
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
