import { useState, useEffect } from "react";
import { CountdownTimer } from "./CountdownTimer";
import { ConfirmButton } from "./ConfirmButton";

interface ActiveReservationProps {
  reservationId: string;
  itemId: string;
  quantity: number;
  expiresAt: string;
  onCleared: () => void;
}

export function ActiveReservation({
  reservationId,
  itemId,
  quantity,
  expiresAt,
  onCleared,
}: ActiveReservationProps) {
  const [expired, setExpired] = useState(false);

  // Auto-clear 2s after expiry so the panel reverts to a clean form.
  useEffect(() => {
    if (!expired) return;
    const t = setTimeout(onCleared, 2000);
    return () => clearTimeout(t);
  }, [expired, onCleared]);

  return (
    <section className="card reservation-card">
      <header className="card-header">
        <h2>Active Reservation</h2>
      </header>
      <dl className="reservation-meta">
        <div>
          <dt>Reservation ID</dt>
          <dd className="mono">{reservationId}</dd>
        </div>
        <div>
          <dt>Item</dt>
          <dd className="mono">{itemId}</dd>
        </div>
        <div>
          <dt>Quantity</dt>
          <dd>{quantity}</dd>
        </div>
      </dl>
      <CountdownTimer expiresAt={expiresAt} onExpire={() => setExpired(true)} />
      <ConfirmButton reservationId={reservationId} disabled={expired} />
      {expired && <p className="error">Reservation expired — please reserve again.</p>}
      <button className="btn-link" onClick={onCleared}>
        Discard
      </button>
    </section>
  );
}
