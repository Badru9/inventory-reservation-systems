import { useState } from "react";
import { Button, Chip } from "@heroui/react";
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

	const press = () => {
		confirm.mutate(reservationId, {
			onSuccess: (data) => {
				setConfirmedAt(data.confirmed_at);
				onConfirmed?.(data.confirmed_at);
			},
		});
	};

	if (confirmedAt) {
		return (
			<Chip color="success" variant="soft">
				<strong>Confirmed</strong> at {new Date(confirmedAt).toLocaleTimeString()}
			</Chip>
		);
	}

	return (
		<div className="flex flex-col gap-2">
			<Button fullWidth isDisabled={disabled} isPending={confirm.isPending} onPress={press}>
				{confirm.isPending ? "Confirming" : "Confirm Purchase"}
			</Button>
			{confirm.isError && (
				<p className="text-danger text-sm">Error: {getErrorMessage(confirm.error)}</p>
			)}
		</div>
	);
}