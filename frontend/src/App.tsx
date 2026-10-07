import { useEffect, useState } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Card, Chip } from "@heroui/react";
import { StockCard } from "./components/StockCard";
import { ReservationForm } from "./components/ReservationForm";
import { ActiveReservation } from "./components/ActiveReservation";

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
		<main className="mx-auto max-w-[1100px] px-6 py-8">
			<header className="mb-6">
				<div className="flex items-center gap-3">
					<h1 className="text-2xl font-bold tracking-tight">INDICO Flash-Sale Dashboard</h1>
					<Chip color="accent" size="sm" variant="soft">
						live demo
					</Chip>
				</div>
				<p className="text-sm text-muted">High-concurrency inventory reservation demo · Go + Gin + Postgres</p>
			</header>

			<div className="grid grid-cols-1 gap-5 md:grid-cols-2">
				<div className="flex flex-col gap-5">
					<StockCard itemId={trackedItem} />
					{!active && (
						<ReservationForm
							defaultItemId={trackedItem}
							onReserved={(r) => setActive(r)}
						/>
					)}
				</div>
				<div className="flex flex-col gap-5">
					{active ? (
						<ActiveReservation
							reservationId={active.reservation_id}
							itemId={active.item_id}
							quantity={active.quantity}
							expiresAt={active.expires_at}
							onCleared={() => setActive(null)}
						/>
					) : (
						<Card className="min-h-[260px] items-center justify-center" variant="secondary">
							<Card.Content>
								<p className="text-muted">No active reservation. Use the form on the left to start one.</p>
							</Card.Content>
						</Card>
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