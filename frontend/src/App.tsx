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

function App() {
  const [products, setProducts] = useState<Product[]>([]);

  useEffect(() => {
    const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "";

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
  }, []);

  return (
    <main>
      <h1>Nomos MVP</h1>
      <p>Produktkatalog</p>
      <ul>
        {products.map((product) => (
          <li key={product.id}>{product.name}</li>
        ))}
      </ul>
    </main>
  );
}

export default App;
