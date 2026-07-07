import { apiFetch } from "./client";
import type { Event, EventDetail, Venue } from "@/lib/types";

export interface EventFilters {
  genre?: string;
  location?: string;
}

export function listEvents(filters: EventFilters = {}) {
  const params = new URLSearchParams();
  if (filters.genre) params.set("genre", filters.genre);
  if (filters.location) params.set("location", filters.location);
  const qs = params.toString();
  return apiFetch<{ events: Event[] }>(`/events${qs ? `?${qs}` : ""}`, {
    auth: false,
  });
}

export function getEvent(id: string) {
  return apiFetch<EventDetail>(`/events/${id}`, { auth: false });
}

export function createEvent(input: {
  title: string;
  description: string;
  genre: string;
  location: string;
  venue_id: string;
  starts_at: string;
}) {
  return apiFetch<Event>("/events", { method: "POST", body: input });
}

export function publishEvent(id: string) {
  return apiFetch<void>(`/events/${id}/publish`, { method: "POST" });
}

export function setPricing(
  eventId: string,
  input: { sector_id: string; amount: number; currency?: string },
) {
  return apiFetch(`/events/${eventId}/pricing`, {
    method: "POST",
    body: input,
  });
}

export function createVenue(input: { name: string; address: string }) {
  return apiFetch<Venue>("/venues", { method: "POST", body: input });
}

export function getVenue(id: string) {
  return apiFetch<Venue>(`/venues/${id}`);
}

export function addSector(
  venueId: string,
  input: { name: string; row_count: number; col_count: number },
) {
  return apiFetch(`/venues/${venueId}/sectors`, {
    method: "POST",
    body: input,
  });
}
