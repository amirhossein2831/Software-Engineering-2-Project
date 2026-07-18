import { apiFetch } from "./client";
import type { Event, EventDetail, Venue, Sector } from "@/lib/types";

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

export function listAllEvents() {
  return apiFetch<{ events: Event[] }>("/events?include_drafts=true", {
    auth: false,
  });
}

export function updateEvent(
  id: string,
  input: {
    title: string;
    description: string;
    genre: string;
    location: string;
    starts_at: string;
  },
) {
  return apiFetch<Event>(`/events/${id}`, { method: "PATCH", body: input });
}

export function deleteEvent(id: string) {
  return apiFetch<void>(`/events/${id}`, { method: "DELETE" });
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

export function listVenues() {
  return apiFetch<{ venues: Venue[] }>("/venues", { auth: false });
}

export function createVenue(input: { name: string; address: string }) {
  return apiFetch<Venue>("/venues", { method: "POST", body: input });
}

export function getVenue(id: string) {
  return apiFetch<Venue>(`/venues/${id}`);
}

export function updateVenue(
  id: string,
  input: { name: string; address: string },
) {
  return apiFetch<Venue>(`/venues/${id}`, { method: "PATCH", body: input });
}

export function deleteVenue(id: string) {
  return apiFetch<void>(`/venues/${id}`, { method: "DELETE" });
}

export function addSector(
  venueId: string,
  input: { name: string; row_count: number; col_count: number },
) {
  return apiFetch<Sector>(`/venues/${venueId}/sectors`, {
    method: "POST",
    body: input,
  });
}

export function deleteSector(venueId: string, sectorId: string) {
  return apiFetch<void>(`/venues/${venueId}/sectors/${sectorId}`, {
    method: "DELETE",
  });
}
