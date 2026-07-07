# Event Ticketing Platform — Technical Design

**Date:** 2026-07-05
**Course:** Software Engineering II — Design and Architecture of an End-to-End Event Ticketing Platform
**Scope of this document:** the technical design — stack, microservice decomposition, core booking/concurrency flow, integrations, and deployment topology. UML diagrams and the written docs (Risk Analysis, Product Vision, Jira) already exist and will be reconciled against this design; they are out of scope here.

---

## 1. Goals (from the project definition)

- Scalable, highly available ticketing platform: event creation → real-time seat selection → secure payment → ticket issuance.
- **No double-booking** under high concurrency; graceful behavior during flash-crowd traffic spikes.
- Bullet-proof checkout: failed/timed-out transactions release locked seats automatically (no "locked limbo").
- Decoupled domains that scale independently.
- One-command local spin-up with no manual troubleshooting (environment portability).

---

## 2. Technology Stack

### Client
- **Next.js + React (TypeScript)** — SSR for event discovery/SEO, rich real-time UI.
- **TanStack Query** for server state; **Zustand** for local UI state.
- **WebSocket** client for live seat availability + order status.
- **SVG/Canvas** seat-map rendering with live Available / Locked / Booked states.
- Dedicated **waiting-room** page that holds users before the booking flow.

### Backend (all services)
- **Go** with **Fiber v3** (HTTP/REST at the edge) and **gRPC** for internal service-to-service calls.
- **GORM** for persistence.
- Primary decoupling is **asynchronous events over Kafka**; gRPC is used only where a synchronous call is genuinely required.

### Data
- **PostgreSQL** — single shared instance, **one database per service** (no shared tables, no cross-service joins). Source of truth for identity, catalog, seat inventory, orders/billing, tickets.
- **Redis** — seat locks (`SET NX PX` + TTL), waiting-room queue + admission tokens, rate-limit counters, hot catalog cache.

### Messaging
- **Kafka (KRaft mode, no Zookeeper)** — durable event backbone: order/seat/payment events, notification fan-out, and the analytics stream.

### Cross-cutting
- **Realtime Gateway:** Go WebSocket hub subscribing to Kafka, pushing seat/order updates to browsers.
- **Saga:** orchestration-based, owned by the Checkout service, with explicit compensation.
- **Observability:** Prometheus + Grafana (metrics), OpenTelemetry → Jaeger (distributed tracing across the checkout saga).

### DevOps
- **Local:** Docker Compose — one command, hermetic (all Go services + Postgres + Redis + Kafka + Mailhog + Mock Payment + Next.js).
- **Prod:** Kubernetes (Helm chart per service, HPA autoscaling on the hot path), **Terraform** for cloud infra, GitLab CI/CD (already running self-hosted).

---

## 3. Project Conventions

- **Direct domain = DB mapping.** Services are small; we do **not** split domain models from persistence models. **GORM structs are the domain models**, mapped directly to tables. No hexagonal/onion mapping ceremony, no anemic DTO layers.
  - **Single caveat:** each event-producing service has a thin **transactional outbox** table. This exists purely for reliable Kafka publishing (atomic with the state change), not to introduce a domain/persistence split.
- **Database-per-service.** Services never touch another service's tables. Cross-service facts travel as Kafka events or gRPC calls.
- **REST at the edge, gRPC inside.** Clients speak REST/JSON + WebSocket to the Gateway; services speak gRPC to each other.
- **Idempotency everywhere.** Idempotency keys on write paths; consumers dedupe by event id (Kafka is at-least-once).
- **Integrations behind interfaces.** Payment and notification senders are Go interfaces with a local default implementation and an optional real adapter selected by env var.

---

## 4. Service Decomposition (Bounded Contexts)

| # | Service (Go / Fiber) | Owns | Storage | Key inbound | Key outbound (Kafka) |
|---|---|---|---|---|---|
| 1 | **API Gateway** | Edge routing, JWT verify, rate limit, waiting-room admission check | Redis (counters) | REST (client) | gRPC to services |
| 2 | **Identity & Access (IAM)** | Users, register/login, JWT issue + refresh, roles (buyer / organizer / admin) | Postgres | REST + gRPC | `user.registered` |
| 3 | **Event Catalog** | Events, venues/halls, sectors, seat-map layout, search/filter | Postgres + Redis cache | REST + gRPC | `event.published`, `seatmap.created` |
| 4 | **Reservation & Inventory** ⭐ | Seat state machine (Available→Locked→Booked), Redis TTL locks, holds | Postgres (truth) + Redis (locks) | gRPC `LockSeats` / `Release` / `Commit` | `seat.locked`, `seat.released`, `seat.booked` |
| 5 | **Checkout & Billing** | Order lifecycle, **saga orchestrator**, external payment, compensation | Postgres | gRPC `CreateOrder` | `order.created`, `payment.succeeded`, `payment.failed` |
| 6 | **Ticketing** | Issues final ticket + QR hash after payment | Postgres | Kafka `payment.succeeded` | `ticket.issued` |
| 7 | **Notification** | SMS/Email on booking/payment/schedule changes | Postgres (outbox) | Kafka (many) | — |
| 8 | **Realtime Gateway** | WebSocket hub → live seat + order status | (stateless) | Kafka (seat/order) | WS to client |
| 9 | **Waiting Room** | Queue tokens, admission throttling during drops | Redis | REST | admission tokens |

**Bonus (Section 5 of the brief):** **Analytics** service — consumes the Kafka stream for organizer dashboards (sales, revenue, remaining capacity).

### DDD boundary rationale
- **Live seat availability lives in Reservation, not Catalog.** Catalog serves the *static* seat map; Reservation owns *live* seat state. This resolves the DDD boundary question in the brief and removes the circular dependency between browsing and buying.
- **Ticketing is split from Checkout** so payment logic and ticket issuance fail/scale independently; Ticketing simply reacts to `payment.succeeded`.
- **Waiting Room is separate from the Gateway** so throttling can be reasoned about and load-tested in isolation, though they cooperate (Gateway validates the admission token).

---

## 5. Core Booking Flow — Concurrency + Checkout Saga

### 5.1 Seat locking (Reservation service)
- Live lock in Redis: `SET lock:seat:{eventId}:{seatId} {holdId} NX PX 600000` → **10-minute TTL** (configurable; matches the brief's user story).
- **All-or-nothing multi-seat lock** via a single **Lua script**: every requested seat is acquired or none are — no partial holds, no deadlock.
- On success: write a `Reservation` row (`status=Locked, expires_at`) in Postgres + emit `seat.locked`.
- **Double-booking backstop:** Postgres **partial unique index** on `(event_id, seat_id)` where `status = Booked`. Even if Redis is lost, the database physically cannot commit the same seat twice.
- **Expiry reaper:** a background sweeper releases `Locked` rows past `expires_at` (→ Available, emit `seat.released`). Redis TTL is the fast path; the sweeper is the durable backstop. We do **not** rely on fragile Redis keyspace notifications.

### 5.2 Checkout saga (orchestration-based, Checkout service)
1. Client holding locked seats → `POST /checkout` → Gateway → `Checkout.CreateOrder` (gRPC).
2. Checkout re-validates the hold via gRPC to Reservation → creates `Order(status=Pending)` → emits `order.created`.
3. Checkout calls the payment gateway (interface — mock or real).
4. **Success:** `Order → Paid`; gRPC `Reservation.Commit` flips seats Locked→Booked and deletes the Redis locks; emit `payment.succeeded` + `seat.booked`.
5. **Ticketing** consumes `payment.succeeded` → generates ticket + QR → emit `ticket.issued`.
6. **Notification / Realtime / Analytics** consume downstream — no coupling back to Checkout.

### 5.3 Compensation (rollback) paths
- **Payment fail/timeout** → `Order → Failed`; gRPC `Reservation.Release` → seats Available; emit `payment.failed`; Notification informs the user.
- **Hold expired at commit** → Reservation rejects Commit → Checkout aborts (refund if already charged).
- **User abandons** → Redis TTL + sweeper auto-release. No stuck "locked limbo."

### 5.4 Reliability primitives
- **Transactional outbox:** the state change and its event row are committed in the same Postgres transaction; a relay publishes to Kafka → no lost or ghost events.
- **Idempotency keys** on checkout + payment → retries never double-charge.
- **Idempotent consumers** (dedupe by event id) → at-least-once delivery is safe.

### 5.5 Waiting Room admission
- Redis **sorted set** = the queue; admission tokens granted at a controlled rate.
- The Gateway rejects booking calls without a valid admission token → protects Reservation/Checkout from flash-crowd collapse. Overflow surfaces as a graceful queue position, not a crash.

---

## 6. External Integrations

### Notification (pluggable senders behind a Go interface)
- **Email:** real SMTP driver configured entirely by env (`SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM`). Local dev points at a **Mailhog** container (catches all mail; web UI on :8025) — offline, no accounts.
- **SMS:** same interface; default **log/stub** driver (writes to outbox + logs); optional Twilio adapter via env.

### Payment (one gRPC interface, two implementations)
- **Default (local, offline): `MockPaymentGateway`** — small internal Go service with configurable success/failure/latency (also the **load/stress-test lever** for simulating gateway timeouts).
- **Optional (real): `StripeAdapter`** — Stripe **test mode**, enabled by `PAYMENT_PROVIDER=stripe` + test keys.
- The checkout saga only knows the interface, so compensation behaves identically for mock and real. Real Stripe is never on the critical path for `docker compose up`.

---

## 7. Deployment / Runtime Topology

### Local (Docker Compose — one command, hermetic)
All Go services + **one Postgres** (database-per-service) + Redis + Kafka (KRaft) + Mailhog + Mock Payment + Next.js frontend. No external dependencies required to boot.

### Production (Kubernetes)
- **Helm chart per service**; **HPA autoscaling** focused on the Reservation/Checkout hot path.
- **Terraform** provisions cloud infra (K8s cluster, managed Postgres, load balancer).
- **Prometheus + Grafana** metrics; **Jaeger** traces; NGINX Ingress for TLS + routing.

---

## 8. Kafka Topics (event catalog)

| Topic | Producer | Primary consumers |
|---|---|---|
| `user.registered` | IAM | Notification, Analytics |
| `event.published` | Catalog | Analytics, (cache warmers) |
| `seat.locked` / `seat.released` / `seat.booked` | Reservation | Realtime, Analytics |
| `order.created` | Checkout | Analytics |
| `payment.succeeded` / `payment.failed` | Checkout | Ticketing, Notification, Realtime, Analytics |
| `ticket.issued` | Ticketing | Notification, Realtime |

All topics carry an event id for consumer-side idempotency.

---

## 9. Out of Scope / Phasing

- This document defines the target architecture. Implementation proceeds service-by-service (each with its own plan), starting with the vertical booking slice: **Catalog → Reservation → Checkout → Ticketing**, plus the Gateway and IAM needed to exercise it end-to-end.
- Analytics is the graded bonus and is built after the core slice is green.
- UML diagrams and the Risk Analysis / Product Vision docs already exist and will be checked against this design in a later pass.
