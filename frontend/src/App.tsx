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

function App() {
  const [products, setProducts] = useState<Product[]>([]);
  const [selectedProduct, setSelectedProduct] = useState<ProductDetail | null>(null);
  const [requirements, setRequirements] = useState<string[]>([]);
  const [requirementsFilter, setRequirementsFilter] = useState("");
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
      setRequirements([]);
      setRequirementsFilter("");
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

  const visibleRequirements = useMemo(() => {
    const source = requirements.length > 0 ? requirements : selectedProduct?.requirements ?? [];

    if (requirementsFilter.trim() === "") {
      return source;
    }

    const query = requirementsFilter.toLowerCase();
    return source.filter((item) => item.toLowerCase().includes(query));
  }, [requirements, requirementsFilter, selectedProduct]);

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

          <h3>Rules</h3>
          <ul>
            {(selectedProduct.rules ?? []).map((rule) => (
              <li key={rule}>{rule}</li>
            ))}
          </ul>

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
