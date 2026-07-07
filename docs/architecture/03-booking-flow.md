# Core Booking Flow — Concurrency & Checkout Saga

**Project:** End-to-End Event Ticketing Platform
**Date:** 2026-07-05

This is the heart of the platform: how seats are locked safely under high
concurrency, how the multi-step checkout is orchestrated as a saga, and how
every failure path releases seats cleanly. See
[`02-services.md`](02-services.md) for the surrounding service APIs.

---

## 1. Seat State Machine

A seat (per event) lives in exactly one state:

```
        LockSeats (Redis NX ok)          Commit (payment ok)
 AVAILABLE ──────────────────▶ LOCKED ──────────────────▶ BOOKED
     ▲                           │                          (terminal)
     │        Release /          │
     └────── TTL expiry ─────────┘
             (reaper)
```

- **AVAILABLE → LOCKED:** a buyer acquires a hold. Fast path is a Redis lock; the durable record is a `SeatState` row + a `Reservation` (hold).
- **LOCKED → BOOKED:** payment succeeded; the seat is permanently sold.
- **LOCKED → AVAILABLE:** the hold is released — explicitly (cancel/compensation) or automatically (TTL expiry via the reaper).
- **BOOKED is terminal** and protected by a Postgres partial-unique index.

---

## 2. Seat Locking (Reservation service)

### 2.1 Redis lock, all-or-nothing
A buyer usually wants several seats at once. Partial locks (some seats yes, some
no) are unacceptable and naive multi-key locking risks deadlock. We acquire all
requested seats atomically with a single **Lua script**:

```
-- KEYS = seat lock keys, ARGV[1] = holdId, ARGV[2] = ttlMillis
for i, key in ipairs(KEYS) do
  if redis.call('GET', key) then return i end   -- someone holds seat i → abort
end
for i, key in ipairs(KEYS) do
  redis.call('SET', key, ARGV[1], 'PX', tonumber(ARGV[2]))
end
return 0                                          -- 0 = all acquired
```

The script runs atomically, so no interleaving between the check and the set.
Return value `0` = success; any `i > 0` = seat `i` was taken → **nothing is
locked**, caller retries with a different selection.

### 2.2 Durable record
On a successful lock, in one Postgres transaction:
1. upsert each `SeatState` to `locked` with the `hold_id`,
2. insert a `Reservation{ status=locked, expires_at = now + TTL }`,
3. insert `seat.locked` into `outbox_events`.

### 2.3 Double-booking backstop (defense in depth)
`SeatState` carries a **partial unique index**:
```
CREATE UNIQUE INDEX uniq_booked_seat
  ON seat_states (event_id, seat_id) WHERE status = 'booked';
```
Redis is the fast mutual-exclusion layer; this index is the last line of
defense. Even if Redis is flushed or a bug double-commits, the database
physically rejects the second `booked` write.

### 2.4 Hold TTL & the reaper
- TTL default **10 minutes** (configurable), matching the user story
  "keep my seat locked for 10 minutes."
- The Redis key auto-expires (fast path). Independently, a **reaper** goroutine
  scans `Reservation` rows where `status=locked AND expires_at < now()`, sets the
  seats back to `available`, marks the hold `released`, and emits `seat.released`.
- We rely on the reaper — **not** Redis keyspace notifications (which are
  best-effort and easily missed) — for the authoritative release.

---

## 3. Checkout Saga (orchestration-based, Checkout service)

The saga is orchestrated (a central coordinator drives each step) rather than
choreographed, because the checkout has a strict order and needs clean
compensation.

### 3.1 Happy path

```
Client                Gateway        Checkout           Reservation      Payment        Kafka
  │  POST /checkout ─────▶│               │                  │              │             │
  │  (idempotency key)    │── CreateOrder─▶│                  │              │             │
  │                       │               │─ ConfirmHold ───▶│              │             │
  │                       │               │◀──── ok ─────────│              │             │
  │                       │               │  Order=Pending   │              │             │
  │                       │               │────────────────  order.created ─────────────▶ │
  │                       │               │─ Charge ─────────────────────▶ │             │
  │                       │               │◀──── success ─────────────────│              │
  │                       │               │  Order=Paid      │              │             │
  │                       │               │─ Commit ────────▶│              │             │
  │                       │               │◀──── ok ─────────│ (locked→booked, del locks) │
  │                       │               │──────────  payment.succeeded, seat.booked ──▶ │
  │◀── 200 {order:paid} ──│◀──────────────│                  │              │             │
```

Downstream, off the same Kafka events:
- **Ticketing** consumes `payment.succeeded` → issues QR ticket → `ticket.issued`.
- **Notification** → sends confirmation email/SMS.
- **Realtime** → pushes "order confirmed" + booked seats to browsers.
- **Analytics** → updates dashboards.

### 3.2 Saga steps & compensation

| # | Step (forward) | Compensation (on later failure) |
|---|---|---|
| 1 | `ConfirmHold` — hold still valid? | none (read-only) |
| 2 | Create `Order=Pending`, emit `order.created` | mark `Order=Compensated` |
| 3 | `Charge` payment | `Refund` (if charge succeeded but a later step fails) |
| 4 | `Commit` seats (locked→booked) | `Release` seats back to available |

### 3.3 Failure paths (each leaves the system consistent)

- **Hold expired before checkout** — `ConfirmHold` fails → Order never created →
  client told to reselect. Seats already free (reaper).
- **Payment declined / times out** — `Order=Failed`; `Reservation.Release` frees
  the seats; emit `payment.failed`; Notification tells the buyer. This is the
  primary postmortem scenario in the brief ("payment gateway downtime leaving
  seats in locked limbo") — the compensation is exactly what prevents the limbo.
- **Commit fails after a successful charge** (e.g., hold expired in the race
  window) — `Refund` the payment, `Order=Compensated`, seats released.
- **Client/network drops mid-saga** — the saga is server-driven and idempotent;
  a retried `POST /checkout` with the same idempotency key returns the existing
  order rather than starting a second one.

---

## 4. Reliability Primitives

### 4.1 Transactional outbox
A state change and its event are committed in the **same Postgres transaction**
(write to the table + insert into `outbox_events`). A relay process reads
unpublished rows and pushes them to Kafka, marking them published. This
guarantees the event fires **iff** the state change committed — no lost events,
no ghost events on rollback.

### 4.2 Idempotency
- **Write endpoints** (`POST /checkout`) require an idempotency key stored
  unique on `Order`; replays return the original result.
- **Consumers** dedupe by `event_id` (a processed-events table or key), so
  Kafka's at-least-once delivery never double-issues a ticket or double-sends a mail.

### 4.3 Ordering & keys
Kafka messages are keyed by `event_id` (seat topics) / `order_id` (order topics)
so all updates for one event/order land on the same partition and stay ordered.

---

## 5. Virtual Waiting Room (traffic throttling)

When a popular event drops, unfiltered traffic would overwhelm Reservation and
Checkout. The Waiting Room absorbs the spike:

1. Client hits a gated event → routed to `POST /queue/{eventId}/join`.
2. Waiting Room adds the user to a Redis **sorted set** (score = enqueue time)
   and returns a position + poll token.
3. Admission is drained at a **configurable rate**; when the user reaches the
   front, `GET /queue/status` returns a signed **admission token**.
4. The Gateway requires a valid admission token on booking endpoints; requests
   without one are rejected with a friendly "still in queue" response — never a
   crash or a raw 503.

Under extreme load the queue simply grows and users see honest positions; core
services stay within capacity. (The brief's bonus postmortem — "queue worker
fails, users fall back to HTTP 503" — is handled by degrading to a static queue
page rather than passing raw errors to buyers.)

---

## 6. Concurrency Test Strategy (for QA)

- **Load/stress:** thousands of concurrent bots targeting the *same* seats; assert
  exactly one wins each seat and the rest get clean "seat taken" responses.
- **Fault injection:** MockPaymentGateway configured to fail/time out at a set
  rate; assert every failed order releases its seats (no leaked locks).
- **Expiry:** acquire holds, let TTL lapse, assert the reaper frees them and
  `seat.released` fires.
- **Idempotency:** replay `POST /checkout` and duplicate Kafka events; assert no
  double charges and no duplicate tickets.
- **Invariant (mutation-testing target):** across any interleaving, a seat is
  `booked` for at most one order — enforced by Redis lock + the partial-unique index.
