import { useState } from "react";
import { useConfirm } from "../hooks/useReserve";
import { getErrorMessage } from "../api/client";

interface Props {
  reservationId: string;
  disabled?: boolean;
  onConfirmed?: (confirmedAt: string) => void;
}

export function ConfirmButton({ reservationId, disabled, onConfirmed }: Props) {
  const confirm = useConfirm();
  const [confirmedAt, setConfirmedAt] = useState<string | null>(null);

  const click = () => {
    confirm.mutate(reservationId, {
      onSuccess: (data) => {
        setConfirmedAt(data.confirmed_at);
        onConfirmed?.(data.confirmed_at);
      },
    });
  };

  if (confirmedAt) {
    return (
      <div className="confirm-success">
        <strong>Confirmed!</strong> at {new Date(confirmedAt).toLocaleTimeString()}
      </div>
    );
  }

  return (
    <div>
      <button
        className="btn-primary"
        onClick={click}
        disabled={disabled || confirm.isPending}
      >
        {confirm.isPending ? "Confirming…" : "Confirm Purchase"}
      </button>
      {confirm.isError && <p className="error">Error: {getErrorMessage(confirm.error)}</p>}
    </div>
  );
}
