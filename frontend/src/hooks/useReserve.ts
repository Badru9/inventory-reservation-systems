import { useMutation } from "@tanstack/react-query";
import { api, type ReserveResponse, type ConfirmResponse } from "../api/client";

export function useReserve() {
  return useMutation({
    mutationFn: async (input: { user_id: string; item_id: string; quantity: number }) => {
      const res = await api.post<ReserveResponse>("/api/v1/inventory/reserve", input);
      return res.data;
    },
  });
}

export function useConfirm() {
  return useMutation({
    mutationFn: async (reservation_id: string) => {
      const res = await api.post<ConfirmResponse>("/api/v1/inventory/confirm", { reservation_id });
      return res.data;
    },
  });
}
