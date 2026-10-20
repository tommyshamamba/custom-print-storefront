import { useMemo, useState } from "react";

type Product = {
  id: string;
  name: string;
  price: string;
  preview: string;
  tenant: string;
};

const CATALOG: Product[] = [
  {
    id: "die-cut",
    name: "Die Cut Stickers",
    price: "$19",
    preview: "auto-generated",
    tenant: "tommy-stickers",
  },
  {
    id: "holographic",
    name: "Holographic Pack",
    price: "$29",
    preview: "auto-generated",
    tenant: "ai-art-shop",
  },
];

export default function App() {
  const [flexibility, setFlexibility] = useState(true);
  const [status, setStatus] = useState("Storefront ready — FLEXIBILITY");
  const products = useMemo(() => CATALOG, []);

  async function checkout(product: Product) {
    setStatus(`Stripe checkout ${product.name} — multi-tenant ${product.tenant}`);
    try {
      const res = await fetch("http://localhost:8080/api/checkout", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          tenant_id: product.tenant,
          product_id: product.id,
          idempotency_key: `${product.tenant}-${product.id}`,
        }),
      });
      const data = await res.json();
      setStatus(`Queued ${data.order_id} — FLEXIBILITY ${data.flexibility}`);
    } catch {
      setStatus("Go API offline — FLEXIBILITY true local preview");
    }
  }

  return (
    <main style={{ fontFamily: "Arial", padding: 24, maxWidth: 900, margin: "0 auto" }}>
      <h1>Sticker Mule Clone — React storefront — FLEXIBILITY</h1>
      <p>
        Multi-tenant catalog + Stripe checkout + image preview generator. 7M+ users
        5M/day 99.99%. 9+ years Tommy Shamamba.
      </p>
      <button type="button" onClick={() => setFlexibility((v) => !v)}>
        FLEXIBILITY {flexibility ? "ON" : "OFF"}
      </button>
      <ul>
        {products.map((product) => (
          <li key={product.id} style={{ margin: "16px 0" }}>
            <strong>{product.name}</strong> {product.price} — tenant {product.tenant} —
            preview {product.preview}
            <div>
              <button type="button" onClick={() => checkout(product)}>
                Stripe Checkout FLEXIBILITY
              </button>
            </div>
          </li>
        ))}
      </ul>
      <p>{status}</p>
      <p>
        Same repo as Expo mobile — single codebase web+mobile FLEXIBILITY true —
        Go GraphQL Redis pub/sub Stripe.
      </p>
    </main>
  );
}
