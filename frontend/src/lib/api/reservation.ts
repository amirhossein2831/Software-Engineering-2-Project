import { apiFetch } from "./client";
import type { Reservation, SeatState } from "@/lib/types";

export function seatStates(eventId: string) {
  return apiFetch<{ seats: SeatState[] }>(`/events/${eventId}/seats`, {
    auth: false,
  });
}

export function hold(
  eventId: string,
  seatIds: string[],
  admissionToken?: string,
) {
  return apiFetch<Reservation>("/holds", {
    method: "POST",
    body: { event_id: eventId, seat_ids: seatIds },
    admissionToken,
  });
}

export function releaseHold(holdId: string) {
  return apiFetch<Reservation>(`/holds/${holdId}/release`, { method: "POST" });
}
