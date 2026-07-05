# Service Decomposition (Bounded Contexts)

**Project:** End-to-End Event Ticketing Platform
**Date:** 2026-07-05

Detailed design of each service: responsibility, data model (GORM entities),
external API (REST) and internal API (gRPC), and the events it produces and
consumes. See [`01-high-level-architecture.md`](01-high-level-architecture.md)
for the big picture and [`03-booking-flow.md`](03-booking-flow.md) for the
concurrency/saga internals.

**Conventions applied to every service:**
- Own Postgres database; no cross-service tables.
- GORM structs are the domain models (direct mapping).
- Event producers include a **transactional outbox** table (`outbox_events`).
- All write endpoints accept an idempotency key; all consumers dedupe by `event_id`.

---

## 1. API Gateway

**Responsibility:** the single edge for the browser. Terminates REST/JSON,
verifies JWTs, enforces rate limits, checks the waiting-room admission token,
and routes to internal services over gRPC. Holds no business logic.

- **Storage:** Redis (rate-limit counters; admission-token lookups).
- **External API (REST):** proxies the public routes below (thin passthrough with auth + throttle).
- **Internal:** gRPC clients for IAM, Catalog, Reservation, Checkout.
- **Cross-cutting:** injects the authenticated `user_id`/roles into downstream gRPC metadata; attaches trace context.

---

## 2. Identity & Access (IAM)

**Responsibility:** registration, authentication, JWT lifecycle, and roles.

**Entities (GORM):**
```
User        { ID, Email(unique), PasswordHash, Role(enum: buyer|organizer|admin), CreatedAt }
RefreshToken{ ID, UserID(fk), TokenHash, ExpiresAt, RevokedAt }
```

**External API (REST via Gateway):**
- `POST /auth/register` → create buyer account
- `POST /auth/login` → returns access (short-lived) + refresh token
- `POST /auth/refresh` → rotate access token
- `POST /auth/logout` → revoke refresh token

**Internal API (gRPC):**
- `VerifyToken(access) → {user_id, role}` (used by Gateway; can also be a local JWT verify with IAM's public key)
- `GetUser(user_id) → User`

**Events produced:** `user.registered`.

---

## 3. Event Catalog

**Responsibility:** the source of truth for **static** event and venue data:
events, venues/halls, sectors, and the seat-map **layout**. Serves search and
filtering. Does **not** own live seat availability (that is Reservation's job).

**Entities (GORM):**
```
Venue   { ID, Name, Address, CreatedBy(organizer) }
Sector  { ID, VenueID(fk), Name, RowCount, ColCount }
Event   { ID, Title, Description, Genre, Location, StartsAt, OrganizerID, Status(draft|published) }
Pricing { ID, EventID(fk), SectorID(fk), Price, Currency }
Seat    { ID, EventID(fk), SectorID(fk), RowLabel, Number }   // layout only; state lives in Reservation
```

**External API (REST):**
- `GET /events?genre=&date=&location=&available=` → filtered search (Redis-cached)
- `GET /events/{id}` → event detail + seat-map layout + pricing
- **Organizer/admin:** `POST /venues`, `POST /events`, `POST /events/{id}/publish`, `POST /events/{id}/pricing`

**Internal API (gRPC):**
- `GetEvent(id)`, `GetSeatMap(event_id)`, `GetPricing(event_id, seat_ids)` (called by Reservation/Checkout).

**Events produced:** `event.published`, `seatmap.created`.

---

## 4. Reservation & Inventory ⭐ (concurrency core)

**Responsibility:** owns **live** seat state and enforces that no two buyers
take the same seat. Full internals in [`03-booking-flow.md`](03-booking-flow.md).

**Entities (GORM):**
```
SeatState   { EventID, SeatID, Status(enum: available|locked|booked),
              HoldID(nullable), UpdatedAt,
              // partial unique index (event_id, seat_id) WHERE status='booked' }
Reservation { ID(=HoldID), EventID, UserID, SeatIDs[], Status(locked|committed|released),
              ExpiresAt, CreatedAt }
outbox_events { ID, Topic, Payload, PublishedAt }
```
- **Redis:** `lock:seat:{eventId}:{seatId} = holdId` with `PX` TTL (10 min default).

**Internal API (gRPC):**
- `LockSeats(event_id, user_id, seat_ids[]) → hold_id` (all-or-nothing, Lua)
- `ConfirmHold(hold_id) → ok` (used by Checkout to re-validate before charging)
- `Commit(hold_id) → ok` (locked → booked; on payment success)
- `Release(hold_id) → ok` (compensation / user cancel)

**Events produced:** `seat.locked`, `seat.released`, `seat.booked`.
**Background:** expiry **reaper** releases holds past `ExpiresAt`.

---

## 5. Checkout & Billing

**Responsibility:** order lifecycle and the **saga orchestrator** that
coordinates hold validation, payment, and seat commit, with compensation.

**Entities (GORM):**
```
Order        { ID, UserID, EventID, HoldID, SeatIDs[], Amount, Currency,
               Status(enum: pending|paid|failed|compensated), IdempotencyKey(unique), CreatedAt }
Payment      { ID, OrderID(fk), Provider, ProviderRef, Status, Amount, CreatedAt }
SagaLog      { ID, OrderID, Step, State, CreatedAt }   // audit / recovery
outbox_events{ ... }
```

**External API (REST):**
- `POST /checkout` (idempotency key) → starts the saga → returns order id + status
- `GET /orders/{id}` → order + payment status
- `GET /orders?user=` → buyer's order history

**Internal API (gRPC):** `CreateOrder(...)` (from Gateway).
**Depends on (gRPC):** Reservation (`ConfirmHold`/`Commit`/`Release`), Catalog (`GetPricing`), Payment interface.
**Events produced:** `order.created`, `payment.succeeded`, `payment.failed`.

---

## 6. Ticketing

**Responsibility:** issues the final verifiable ticket with an embedded QR hash
after payment succeeds. Reacts to events only — no synchronous coupling.

**Entities (GORM):**
```
Ticket { ID, OrderID, EventID, SeatID, UserID, QRHash(unique), IssuedAt }
```
- **Consumes:** `payment.succeeded` → generate one ticket per seat, compute `QRHash` (signed), persist.
- **Events produced:** `ticket.issued`.
- **External API (REST):** `GET /tickets/{id}`, `GET /tickets?order=` (QR image rendered client-side from the hash).

---

## 7. Notification

**Responsibility:** transactional email/SMS on booking, payment, and schedule changes.

**Entities (GORM):**
```
NotificationOutbox { ID, Channel(email|sms), To, Template, Payload, Status(pending|sent|failed), CreatedAt }
```
- **Senders (Go interface):** `EmailSender` (SMTP via env; Mailhog locally), `SmsSender` (log-stub default; optional Twilio).
- **Consumes:** `payment.succeeded`, `payment.failed`, `ticket.issued`, `user.registered`, event-schedule changes.
- Writes to its outbox, then a worker delivers (retriable, idempotent by `event_id`).

---

## 8. Realtime Gateway

**Responsibility:** pushes live updates to the browser over WebSocket. Stateless;
horizontally scalable.

- **Consumes (Kafka):** `seat.locked` / `seat.released` / `seat.booked` (per-event fan-out to viewers), `payment.succeeded/failed`, `ticket.issued` (per-user order status).
- **Client protocol:** browser subscribes to `event:{id}` (seat updates) and `user:{id}` (order updates) channels after auth.
- No database; Kafka is the source, WebSocket is the sink.

---

## 9. Waiting Room

**Responsibility:** admission control during flash-crowd drops. Protects
Reservation/Checkout from overload.

- **Storage:** Redis **sorted set** keyed per hot event (`queue:{eventId}`), score = enqueue time.
- **External API (REST):** `POST /queue/{eventId}/join` → position + poll token; `GET /queue/{eventId}/status` → position or an **admission token** once let in.
- **Admission rate** is configurable per event; the Gateway rejects booking calls lacking a valid admission token.

---

## 10. Analytics (bonus — Section 5 of the brief)

**Responsibility:** organizer dashboards — sales progress, revenue, remaining
capacity — built as read-model projections from the Kafka stream.

- **Consumes:** `order.created`, `payment.succeeded`, `seat.booked`, `event.published`.
- **Storage:** its own Postgres read-model tables (denormalized for fast dashboard queries).
- **External API (REST):** `GET /analytics/events/{id}` → live metrics for the owning organizer.

---

## 11. Service Dependency Summary

| Service | Sync deps (gRPC) | Consumes (Kafka) | Produces (Kafka) |
|---|---|---|---|
| API Gateway | IAM, Catalog, Reservation, Checkout, Waiting Room | — | — |
| IAM | — | — | `user.registered` |
| Catalog | — | — | `event.published`, `seatmap.created` |
| Reservation | Catalog | — | `seat.locked/released/booked` |
| Checkout | Reservation, Catalog, Payment | — | `order.created`, `payment.succeeded/failed` |
| Ticketing | — | `payment.succeeded` | `ticket.issued` |
| Notification | — | many | — |
| Realtime | — | seat/order/ticket | — |
| Waiting Room | — | — | — |
| Analytics | — | order/payment/seat/event | — |
