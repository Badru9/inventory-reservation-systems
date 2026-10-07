import { useState } from "react";
import { useReserve } from "../hooks/useReserve";
import { getErrorMessage } from "../api/client";

interface Props {
  defaultItemId: string;
  onReserved: (res: { reservation_id: string; expires_at: string; item_id: string; quantity: number }) => void;
}

export function ReservationForm({ defaultItemId, onReserved }: Props) {
  const [userId, setUserId] = useState("usr_9981");
  const [itemId, setItemId] = useState(defaultItemId);
  const [quantity, setQuantity] = useState(1);
  const reserve = useReserve();

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    reserve.mutate(
      { user_id: userId, item_id: itemId, quantity: Number(quantity) },
      {
        onSuccess: (data) => {
          onReserved({
            reservation_id: data.reservation_id,
            expires_at: data.expires_at,
            item_id: data.item_id,
            quantity: data.quantity,
          });
        },
      },
    );
  };

  return (
    <section className="card">
      <header className="card-header">
        <h2>Reserve Stock</h2>
      </header>
      <form onSubmit={submit} className="form">
        <label>
          <span>User ID</span>
          <input
            type="text"
            value={userId}
            onChange={(e) => setUserId(e.target.value)}
            placeholder="usr_9981"
            required
            minLength={1}
            maxLength={64}
          />
        </label>
        <label>
          <span>Item ID</span>
          <input
            type="text"
            value={itemId}
            onChange={(e) => setItemId(e.target.value)}
            placeholder="item_4021"
            required
            minLength={1}
            maxLength={64}
          />
        </label>
        <label>
          <span>Quantity</span>
          <input
            type="number"
            value={quantity}
            onChange={(e) => setQuantity(Number(e.target.value))}
            min={1}
            max={10000}
            required
          />
        </label>
        <button type="submit" disabled={reserve.isPending}>
          {reserve.isPending ? "Reserving…" : "Reserve"}
        </button>
        {reserve.isError && <p className="error">Error: {getErrorMessage(reserve.error)}</p>}
        {reserve.isSuccess && <p className="success">Reserved {reserve.data.quantity} × {reserve.data.item_id}</p>}
      </form>
    </section>
  );
}
