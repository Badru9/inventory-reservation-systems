import { useEffect, useState } from "react";
import { Label } from "@heroui/react";

interface Props {
	expiresAt: string; // ISO 8601
	onExpire?: () => void;
}

function fmt(ms: number): string {
	if (ms <= 0) return "00:00";
	const total = Math.floor(ms / 1000);
	const m = Math.floor(total / 60);
	const s = total % 60;
	return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

export function CountdownTimer({ expiresAt, onExpire }: Props) {
	const target = new Date(expiresAt).getTime();
	const [remaining, setRemaining] = useState(target - Date.now());

	useEffect(() => {
		const tick = () => {
			const left = target - Date.now();
			setRemaining(left);
			if (left <= 0) {
				onExpire?.();
			}
		};
		tick();
		const id = setInterval(tick, 1000);
		return () => clearInterval(id);
	}, [target, onExpire]);

	const expired = remaining <= 0;
	return (
		<div
			className={`flex items-baseline gap-3 rounded-xl border px-4 py-3 ${
				expired ? "border-danger bg-danger-soft" : "border-border bg-surface-secondary"
			}`}
		>
			<Label className="text-muted text-xs uppercase tracking-wider">
				{expired ? "EXPIRED" : "Expires in"}
			</Label>
			<span
				className={`font-mono text-2xl font-bold tabular-nums ${expired ? "text-danger" : "text-foreground"}`}
			>
				{fmt(remaining)}
			</span>
		</div>
	);
}