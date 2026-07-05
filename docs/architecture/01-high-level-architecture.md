# High-Level Architecture

**Project:** End-to-End Event Ticketing Platform
**Date:** 2026-07-05

This document is the big-picture view. Detailed per-service design is in
[`02-services.md`](02-services.md); the booking/concurrency internals are in
[`03-booking-flow.md`](03-booking-flow.md).

---

## 1. System Context

```
                         ┌───────────────────────────┐
   Buyers / Organizers   │      Next.js + React      │
   (browser)  ─────────▶ │      (SSR web client)     │
                         └────────────┬──────────────┘
                            REST/JSON  │  WebSocket
                                       ▼
                         ┌───────────────────────────┐
                         │        API Gateway        │  ← JWT verify, rate limit,
                         │        (Go / Fiber)       │    waiting-room admission
                         └────────────┬──────────────┘
                                gRPC  │
      ┌───────────────┬──────────────┼───────────────┬────────────────┐
      ▼               ▼              ▼                ▼                ▼
 ┌─────────┐   ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐
 │  IAM    │   │  Catalog   │  │ Reservation│  │  Checkout  │  │Waiting Room│
 └────┬────┘   └─────┬──────┘  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘
      │              │               │               │               │
      └──────────────┴───────┬───────┴───────┬───────┴───────────────┘
                             ▼               ▼
                        ┌─────────┐     ┌─────────┐
                        │ Postgres│     │  Redis  │
                        │ (per-db)│     │ (locks) │
                        └─────────┘     └─────────┘
                             │
                    ─────────┴──────────  Kafka (event backbone)  ──────────┐
                             │                    │                          │
                             ▼                    ▼                          ▼
                       ┌───────────┐        ┌───────────┐            ┌─────────────┐
                       │ Ticketing │        │Notification│           │  Analytics  │
                       └───────────┘        └───────────┘            │   (bonus)   │
                                                                     └─────────────┘
                                             ▲
                       ┌───────────┐         │ push
                       │  Realtime │─────────┘ (WebSocket to browser)
                       │  Gateway  │
                       └───────────┘
```

**Two communication planes:**
- **Synchronous (request/response):** REST at the edge (client → Gateway), gRPC inside (Gateway → services, service → service). Used only where an immediate answer is required (lock seats, create order).
- **Asynchronous (facts):** Kafka. Every meaningful state change is published as an event. This is the primary decoupling mechanism — downstream services (Ticketing, Notification, Realtime, Analytics) never call back into the core.

---

## 2. Architectural Principles

1. **Database-per-service.** Each service owns its schema; no cross-service tables or joins. Facts cross boundaries as events or gRPC calls only.
2. **Direct domain = DB mapping.** Services are small, so GORM structs *are* the domain models. No separate persistence/domain layers. The one exception is a per-service **transactional outbox** table for reliable event publishing.
3. **REST out, gRPC in.** Clients see REST/JSON + WebSocket; internal traffic is gRPC.
4. **Async-first.** If a consumer doesn't need to block the user, it consumes an event instead of being called.
5. **Fail safe, never stuck.** Every seat lock has a TTL; every saga has compensation; nothing stays in "locked limbo."
6. **Idempotent by construction.** Idempotency keys on writes; event ids for consumer dedupe (Kafka is at-least-once).
7. **Hermetic local dev.** `docker compose up` boots the entire platform with no external accounts or network.

---

## 3. Technology Stack (summary)

| Layer | Choice |
|---|---|
| Web client | Next.js + React (TypeScript), TanStack Query, Zustand, WebSocket |
| Edge | NGINX Ingress (prod) + Go/Fiber API Gateway |
| Services | Go, **Fiber v3** (REST), **gRPC** (internal), **GORM** |
| Relational store | PostgreSQL — one shared instance, **database per service** |
| In-memory store | Redis — locks, waiting-room queue, rate limits, cache |
| Event backbone | **Kafka** (KRaft mode, no Zookeeper) |
| Realtime | Go WebSocket hub subscribing to Kafka |
| Payments | gRPC interface → MockPaymentGateway (default) / Stripe test-mode adapter (optional) |
| Email/SMS | SMTP via env (Mailhog locally) / log-stub SMS (optional Twilio) |
| Metrics/Tracing | Prometheus + Grafana; OpenTelemetry → Jaeger |
| Local | Docker Compose | 
| Prod | Kubernetes (Helm per service, HPA), Terraform, GitLab CI/CD |

---

## 4. Domains and Ownership

| Domain | Service | Data owned |
|---|---|---|
| Identity & Access | IAM | users, roles, refresh tokens |
| Event Catalog & Discovery | Catalog | events, venues, sectors, seat-map layout |
| Reservation & Inventory | Reservation | seat live-state, holds/reservations (+ Redis locks) |
| Billing & Checkout | Checkout | orders, payments, saga state |
| Ticketing | Ticketing | issued tickets + QR |
| Notification & Messaging | Notification | notification outbox/log |
| Edge / Admission | API Gateway, Waiting Room | rate-limit + queue state (Redis) |
| Realtime delivery | Realtime Gateway | (stateless) |
| Analytics (bonus) | Analytics | read-model projections |

---

## 5. End-to-End Data Flow (happy path, condensed)

1. Browser loads events (REST → Gateway → Catalog, cached in Redis).
2. On a hot event, the browser is first routed through **Waiting Room**; it receives an admission token when let in.
3. User selects seats → `LockSeats` (Gateway → Reservation). Redis locks acquired all-or-nothing; `seat.locked` published; Realtime pushes the new seat states to every viewer.
4. User pays → `CreateOrder` (Gateway → Checkout). Checkout runs the **saga**: validate hold → charge payment → commit seats.
5. `payment.succeeded` → **Ticketing** issues the QR ticket; **Notification** emails the buyer; **Realtime** flips the order to "confirmed"; **Analytics** updates dashboards.

Failure at any saga step triggers **compensation** (release seats, refund) — detailed in [`03-booking-flow.md`](03-booking-flow.md).

---

## 6. Cross-Cutting Concerns

- **Security:** JWT (access + refresh) issued by IAM, verified at the Gateway; role-based checks (buyer / organizer / admin). TLS at the ingress.
- **Resilience:** seat-lock TTL + reaper, saga compensation, transactional outbox, idempotent consumers, waiting-room throttling, Postgres partial-unique backstop against double-booking.
- **Scalability:** stateless services behind HPA; Reservation/Checkout are the hot path and scale first; Redis and Kafka absorb bursts.
- **Observability:** RED metrics per service (Prometheus), distributed traces across the saga (Jaeger), dashboards + alerts (Grafana) on latency and failed-checkout rate.

---

## 7. Deployment Topology

- **Local:** single `docker-compose.yml` — all Go services, one Postgres (db-per-service), Redis, Kafka (KRaft), Mailhog, Mock Payment, Next.js. One command, offline.
- **Prod:** Kubernetes; Helm chart per service; HPA on the hot path; Terraform provisions the cluster, managed Postgres, and load balancer; Prometheus/Grafana/Jaeger for observability; GitLab CI/CD builds and deploys.
