import { useState, useEffect } from "react";
import { Button, Card, Description, Label } from "@heroui/react";
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

	useEffect(() => {
		if (!expired) return;
		const t = setTimeout(onCleared, 2000);
		return () => clearTimeout(t);
	}, [expired, onCleared]);

	return (
		<Card>
			<Card.Header className="flex-row items-center justify-between">
				<div>
					<Card.Title>Active Reservation</Card.Title>
					<Card.Description>
						Units are locked server-side until expiry or confirmation.
					</Card.Description>
				</div>
				<Button size="sm" variant="ghost" onPress={onCleared}>
					Discard
				</Button>
			</Card.Header>

			<Card.Content className="flex flex-col gap-4">
				<dl className="flex flex-col gap-3">
					<Meta label="Reservation ID" value={reservationId} mono />
					<Meta label="Item" value={itemId} mono />
					<Meta label="Quantity" value={String(quantity)} />
				</dl>

				<CountdownTimer expiresAt={expiresAt} onExpire={() => setExpired(true)} />

				{expired ? (
					<Description className="text-danger">
						Reservation expired — please reserve again.
					</Description>
				) : (
					<ConfirmButton reservationId={reservationId} />
				)}
			</Card.Content>
		</Card>
	);
}

function Meta({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
	return (
		<div className="flex items-baseline justify-between gap-3">
			<Label className="text-muted text-xs uppercase tracking-wider">{label}</Label>
			<dd
				className={`text-sm ${mono ? "font-mono" : ""}`}
				style={{ margin: 0 }}
			>
				{value}
			</dd>
		</div>
	);
}