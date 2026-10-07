import { useState } from "react";
import { Button, Card, Input, Label } from "@heroui/react";
import { useReserve } from "../hooks/useReserve";
import { getErrorMessage } from "../api/client";

interface Props {
	defaultItemId: string;
	onReserved: (res: {
		reservation_id: string;
		expires_at: string;
		item_id: string;
		quantity: number;
	}) => void;
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
		<Card>
			<Card.Header>
				<Card.Title>Reserve Stock</Card.Title>
				<Card.Description>
					Atomically lock units for 5 minutes, then confirm to commit.
				</Card.Description>
			</Card.Header>
			<form onSubmit={submit} className="flex flex-col gap-4">
				<Card.Content className="flex flex-col gap-4">
					<div className="flex flex-col gap-1.5">
						<Label htmlFor="user-id">User ID</Label>
						<Input
							id="user-id"
							maxLength={64}
							minLength={1}
							placeholder="usr_9981"
							required
							value={userId}
							variant="secondary"
							onChange={(e) => setUserId(e.target.value)}
						/>
					</div>

					<div className="flex flex-col gap-1.5">
						<Label htmlFor="item-id">Item ID</Label>
						<Input
							id="item-id"
							maxLength={64}
							minLength={1}
							placeholder="item_4021"
							required
							value={itemId}
							variant="secondary"
							onChange={(e) => setItemId(e.target.value)}
						/>
					</div>

					<div className="flex flex-col gap-1.5">
						<Label htmlFor="qty">Quantity</Label>
						<Input
							id="qty"
							max={10000}
							min={1}
							placeholder="1"
							required
							type="number"
							value={quantity}
							variant="secondary"
							onChange={(e) => setQuantity(Number(e.target.value))}
						/>
					</div>

					{reserve.isError && (
						<p className="text-danger text-sm">Error: {getErrorMessage(reserve.error)}</p>
					)}
					{reserve.isSuccess && (
						<p className="text-success text-sm">
							Reserved {reserve.data.quantity} × {reserve.data.item_id}
						</p>
					)}
				</Card.Content>

				<Card.Footer>
					<Button fullWidth isPending={reserve.isPending} type="submit">
						{reserve.isPending ? "Reserving" : "Reserve"}
					</Button>
				</Card.Footer>
			</form>
		</Card>
	);
}