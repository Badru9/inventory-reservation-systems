import axios, { AxiosError } from "axios";

const baseURL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

export const api = axios.create({
  baseURL,
  timeout: 10_000,
  headers: { "Content-Type": "application/json" },
});

export interface StockView {
  item_id: string;
  total_stock: number;
  reserved_stock: number;
  available_stock: number;
}

export interface ReserveResponse {
  status: string;
  reservation_id: string;
  item_id: string;
  quantity: number;
  expires_at: string;
}

export interface ConfirmResponse {
  status: string;
  reservation_id: string;
  confirmed_at: string;
}

export interface APIErrorEnvelope {
  status: "error";
  error: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
}

export function getErrorMessage(err: unknown): string {
  if (err instanceof AxiosError) {
    const data = err.response?.data as APIErrorEnvelope | undefined;
    if (data?.error) {
      const det = data.error.details;
      if (det && (det.available !== undefined || det.requested !== undefined)) {
        return `${data.error.message} (available=${det.available ?? "?"}, requested=${det.requested ?? "?"})`;
      }
      return data.error.message;
    }
    return err.message;
  }
  if (err instanceof Error) return err.message;
  return "Unknown error";
}
