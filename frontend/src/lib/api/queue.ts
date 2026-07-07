import { apiFetch } from "./client";
import type { QueueStatus } from "@/lib/types";

export function joinQueue(eventId: string) {
  return apiFetch<QueueStatus>(`/queue/${eventId}/join`, { method: "POST" });
}

export function queueStatus(eventId: string) {
  return apiFetch<QueueStatus>(`/queue/${eventId}/status`);
}
