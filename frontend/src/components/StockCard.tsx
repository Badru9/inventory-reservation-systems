import { useState } from "react";
import { useStock } from "../hooks/useStock";
import { getErrorMessage } from "../api/client";

interface Props {
  itemId: string;
}

const KNOWN_ITEMS = [
  { id: "item_4021", name: "Flash Sale Headphones" },
  { id: "item_7777", name: "Limited Sneakers" },
];

export function StockCard({ itemId }: Props) {
  const { data, isLoading, error, refetch, isFetching, dataUpdatedAt } = useStock(itemId);
  const [selected, setSelected] = useState(itemId);

  return (
    <section className="card">
      <header className="card-header">
        <h2>Live Inventory</h2>
        <select
          value={selected}
          onChange={(e) => setSelected(e.target.value)}
          aria-label="Item"
        >
          {KNOWN_ITEMS.map((it) => (
            <option key={it.id} value={it.id}>
              {it.name} ({it.id})
            </option>
          ))}
        </select>
        <button onClick={() => refetch()} disabled={isFetching}>
          {isFetching ? "Refreshing…" : "Refresh"}
        </button>
      </header>

      {isLoading && <p className="muted">Loading…</p>}
      {error && <p className="error">Error: {getErrorMessage(error)}</p>}

      {data && (
        <>
          <div className="stock-grid">
            <Stat label="Total" value={data.total_stock} tone="neutral" />
            <Stat label="Reserved" value={data.reserved_stock} tone="warn" />
            <Stat label="Available" value={data.available_stock} tone="good" />
          </div>
          <footer className="muted small">
            Last update: {new Date(dataUpdatedAt).toLocaleTimeString()} · auto-refresh 3s
          </footer>
        </>
      )}
    </section>
  );
}

function Stat({ label, value, tone }: { label: string; value: number; tone: "neutral" | "warn" | "good" }) {
  return (
    <div className={`stat stat-${tone}`}>
      <div className="stat-value">{value}</div>
      <div className="stat-label">{label}</div>
    </div>
  );
}
