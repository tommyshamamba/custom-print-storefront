import { useState } from "react";

type Product = { id: string; name: string; price: string };

const TENANT = "demo-shop";
const CATALOG: Product[] = [
  { id: "die-cut", name: "Die-cut stickers", price: "$19" },
  { id: "holographic", name: "Holographic pack", price: "$29" },
];

export default function App() {
  const [quantity, setQuantity] = useState(50);
  const [status, setStatus] = useState("");

  async function checkout(product: Product) {
    setStatus(`Submitting ${product.name}...`);
    try {
      const res = await fetch("/api/checkout", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          tenant_id: TENANT,
          product_id: product.id,
          quantity,
          idempotency_key: crypto.randomUUID(),
        }),
      });
      const data = await res.json();
      setStatus(res.ok ? `Order ${data.id}: ${data.status}` : `Error: ${data.error}`);
    } catch {
      setStatus("Checkout API is not reachable. Start the Go server on :8080.");
    }
  }

  return (
    <main style={{ fontFamily: "system-ui", padding: 24, maxWidth: 720, margin: "0 auto" }}>
      <h1>Custom print storefront</h1>
      <label>
        Quantity{" "}
        <input type="number" min={1} max={10000} value={quantity}
          onChange={(e) => setQuantity(Number(e.target.value))} />
      </label>
      <ul>
        {CATALOG.map((p) => (
          <li key={p.id} style={{ margin: "12px 0" }}>
            <strong>{p.name}</strong> {p.price}{" "}
            <button type="button" onClick={() => checkout(p)}>Checkout</button>
          </li>
        ))}
      </ul>
      <p role="status">{status}</p>
    </main>
  );
}
