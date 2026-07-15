export type Role = "buyer" | "organizer" | "admin";

export interface SessionUser {
  id: string;
  email: string;
  role: Role;
}

export interface Tokens {
  access_token: string;
  refresh_token: string;
}

export interface Sector {
  id: string;
  venue_id: string;
  name: string;
  row_count: number;
  col_count: number;
}

export interface Venue {
  id: string;
  name: string;
  address: string;
  sectors?: Sector[];
}

export type EventStatus = "draft" | "published";

export interface Event {
  id: string;
  title: string;
  description: string;
  genre: string;
  location: string;
  venue_id: string;
  organizer_id: string;
  starts_at: string;
  status: EventStatus;
}

export interface Pricing {
  id: string;
  event_id: string;
  sector_id: string;
  amount: number;
  currency: string;
}

export interface Seat {
  id: string;
  event_id: string;
  sector_id: string;
  row_label: string;
  number: number;
}

export interface EventDetail {
  event: Event;
  pricing: Pricing[];
  seats: Seat[];
}

export type SeatStatus = "available" | "locked" | "booked";

export interface SeatState {
  event_id: string;
  seat_id: string;
  status: SeatStatus;
}

export interface Reservation {
  id: string;
  event_id: string;
  user_id: string;
  status: string;
  expires_at: string;
}

export interface HoldDetail {
  reservation: Reservation;
  seat_ids: string[];
}

export type OrderStatus = "pending" | "paid" | "failed" | "compensated";

export interface Order {
  id: string;
  user_id: string;
  event_id: string;
  hold_id: string;
  seat_ids: string[];
  amount: number;
  currency: string;
  status: OrderStatus;
  created_at: string;
}

export interface Ticket {
  id: string;
  order_id: string;
  event_id: string;
  seat_id: string;
  user_id: string;
  qr_hash: string;
  issued_at: string;
}

export interface QueueStatus {
  status: "waiting" | "admitted";
  position?: number;
  admission_token?: string;
}

export interface AdminUser {
  id: string;
  email: string;
  role: Role;
}
