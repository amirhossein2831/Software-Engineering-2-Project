import { apiFetch } from "./client";
import type { Order, Ticket } from "@/lib/types";

export function checkout(
  holdId: string,
  idempotencyKey: string,
  admissionToken?: string,
) {
  return apiFetch<Order>("/checkout", {
    method: "POST",
    body: { hold_id: holdId },
    idempotencyKey,
    admissionToken,
  });
}

export function getOrder(id: string) {
  return apiFetch<{ order: Order; payments: unknown[] }>(`/orders/${id}`);
}

export function listOrders() {
  return apiFetch<{ orders: Order[] }>("/orders");
}

export function listTickets(orderId: string) {
  return apiFetch<{ tickets: Ticket[] }>(`/tickets?order=${orderId}`);
}
