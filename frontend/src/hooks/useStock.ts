import { useQuery } from "@tanstack/react-query";
import { api, type StockView } from "../api/client";

export function useStock(itemId: string, refetchIntervalMs = 3000) {
  return useQuery<StockView>({
    queryKey: ["stock", itemId],
    queryFn: async () => {
      const res = await api.get<{ status: string; data: StockView }>(
        `/api/v1/inventory/stock`,
        { params: { item_id: itemId } },
      );
      return res.data.data;
    },
    enabled: Boolean(itemId),
    refetchInterval: refetchIntervalMs,
    staleTime: 1000,
  });
}
