import { useEffect, useState } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StockCard } from "./components/StockCard";
import { ReservationForm } from "./components/ReservationForm";
import { ActiveReservation } from "./components/ActiveReservation";
import "./App.css";

interface ActiveRes {
  reservation_id: string;
  item_id: string;
  quantity: number;
  expires_at: string;
}

const STORAGE_KEY = "indico.activeReservation";

function Dashboard() {
  const [active, setActive] = useState<ActiveRes | null>(() => {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    try {
      return JSON.parse(raw) as ActiveRes;
    } catch {
      return null;
    }
  });

  useEffect(() => {
    if (active) sessionStorage.setItem(STORAGE_KEY, JSON.stringify(active));
    else sessionStorage.removeItem(STORAGE_KEY);
  }, [active]);

  const trackedItem = active?.item_id ?? "item_4021";

  return (
    <main className="app">
      <header className="app-header">
        <h1>INDICO Flash-Sale Dashboard</h1>
        <p className="muted">High-concurrency inventory reservation demo · Go + Gin + Postgres</p>
      </header>

      <div className="grid">
        <div className="col">
          <StockCard itemId={trackedItem} />
          {!active && (
            <ReservationForm
              defaultItemId={trackedItem}
              onReserved={(r) => setActive(r)}
            />
          )}
        </div>
        <div className="col">
          {active ? (
            <ActiveReservation
              reservationId={active.reservation_id}
              itemId={active.item_id}
              quantity={active.quantity}
              expiresAt={active.expires_at}
              onCleared={() => setActive(null)}
            />
          ) : (
            <section className="card placeholder">
              <p className="muted">No active reservation. Use the form on the left to start one.</p>
            </section>
          )}
        </div>
      </div>
    </main>
  );
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Dashboard />
    </QueryClientProvider>
  );
}
