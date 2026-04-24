import { useEffect, useState } from "react";

type Product = {
  id: string;
  name: string;
  version: string;
};

type ProductResponse = {
  items: Product[];
  count: number;
};

type ProductDetail = {
  id: string;
  name: string;
  version: string;
  beschreibung?: string;
};

function App() {
  const [products, setProducts] = useState<Product[]>([]);
  const [selectedProduct, setSelectedProduct] = useState<ProductDetail | null>(null);
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
        // In Dev ohne laufendes Backend soll die UI weiterhin rendern.
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
    } catch {
      // Fehler im Detail-Load sollen die Seite nicht blockieren.
    }
  };

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
          <p>Version: {selectedProduct.version}</p>
          <p>{selectedProduct.beschreibung}</p>
        </section>
      ) : null}
    </main>
  );
}

export default App;
