import { useState } from "react";
import { Button, Card, Label, ListBox, Select, Spinner } from "@heroui/react";
import { useStock } from "../hooks/useStock";
import { getErrorMessage } from "../api/client";

interface Props {
  itemId: string;
}

const KNOWN_ITEMS = [
  { id: "item_4021", name: "Flash Sale Headphones" },
  { id: "item_7777", name: "Limited Sneakers" },
] as const;

export function StockCard({ itemId }: Props) {
  const { data, isLoading, error, refetch, isFetching, dataUpdatedAt } =
    useStock(itemId);
  const [selected, setSelected] = useState<string>(itemId);
  const selectedItem = KNOWN_ITEMS.find((it) => it.id === selected);

  return (
    <Card>
      <Card.Header className="flex flex-col items-start gap-3">
        <div className="flex flex-1 flex-col">
          <Card.Title>Live Inventory</Card.Title>
          <Card.Description>
            Real-time stock for the selected item
          </Card.Description>
        </div>
        <div className="flex flex-row w-full items-center gap-2">
          <Select
            fullWidth
            aria-label="Item"
            value={selected}
            onChange={(value) => setSelected(String(value))}
          >
            <Select.Trigger>
              <Select.Value />
              <Select.Indicator />
            </Select.Trigger>
            <Select.Popover>
              <ListBox>
                {KNOWN_ITEMS.map((it) => (
                  <ListBox.Item
                    key={it.id}
                    id={it.id}
                    textValue={`${it.name} (${it.id})`}
                  >
                    {it.name}{" "}
                    <span className="text-muted text-xs">({it.id})</span>
                    <ListBox.ItemIndicator />
                  </ListBox.Item>
                ))}
              </ListBox>
            </Select.Popover>
          </Select>
          <Button
            isPending={isFetching}
            variant="secondary"
            onPress={() => refetch()}
          >
            {isFetching ? "Refreshing" : "Refresh"}
          </Button>
        </div>
      </Card.Header>

      <Card.Content className="flex flex-col gap-3">
        {isLoading && (
          <div className="flex items-center gap-2 text-muted">
            <Spinner size="sm" /> <Label>Loading stock…</Label>
          </div>
        )}
        {error && (
          <p className="text-danger text-sm">Error: {getErrorMessage(error)}</p>
        )}

        {data && (
          <>
            <div className="grid grid-cols-3 gap-3">
              <Stat label="Total" value={data.total_stock} tone="neutral" />
              <Stat label="Reserved" value={data.reserved_stock} tone="warn" />
              <Stat
                label="Available"
                value={data.available_stock}
                tone="good"
              />
            </div>
            <p className="text-muted text-xs">
              {selectedItem?.name ?? selected} · last update{" "}
              {new Date(dataUpdatedAt).toLocaleTimeString()} · auto-refresh 3s
            </p>
          </>
        )}
      </Card.Content>
    </Card>
  );
}

function Stat({
  label,
  value,
  tone,
}: {
  label: string;
  value: number;
  tone: "neutral" | "warn" | "good";
}) {
  const toneClass =
    tone === "good"
      ? "text-success"
      : tone === "warn"
        ? "text-warning"
        : "text-foreground";
  return (
    <div className="rounded-xl border border-border bg-surface-secondary p-4 text-center">
      <div className={`text-3xl font-bold leading-none ${toneClass}`}>
        {value}
      </div>
      <div className="text-muted mt-2 text-xs uppercase tracking-wider">
        {label}
      </div>
    </div>
  );
}
