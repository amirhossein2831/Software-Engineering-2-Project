# Web Client (Next.js)

**Project:** End-to-End Event Ticketing Platform
**Date:** 2026-07-06

The browser client for buyers, organizers, and admins. It is the only piece that
talks to the **API Gateway** (`:8080`) — never to individual services — over REST
+ WebSocket. See [`02-services.md`](02-services.md) for the surfaces it consumes
and the role model it honors.

---

## 1. One project, role-separated areas

A **single Next.js app** (not two) serves all three roles. Buyer and staff UIs
share auth, the API client, TypeScript types, and the design system; splitting
them into separate projects would duplicate all of that. Separation is by
**App Router route groups**, not repositories:

```
src/app/
  (shop)/        buyer-facing: browse, seat selection, checkout, tickets
  admin/         organizer + admin: dashboard, event/venue management, users
```

- **buyer** — everything under `(shop)`.
- **organizer** — `(shop)` + `/admin` scoped to their own events/venues/analytics.
- **admin** — `/admin` unscoped + user management (`/admin/users`).

Role gating is enforced **server-side by the Gateway** (JWT → `X-User-Role`); the
client mirrors it for UX only (hides/redirects), never as the security boundary.

---

## 2. Stack & conventions

| Concern | Choice |
|---|---|
| Framework | Next.js 15 App Router, React 19, TypeScript (strict) |
| Styling | Tailwind CSS v4 (`@theme` tokens in `globals.css`), no config file |
| Server state | TanStack Query (fetch, cache, invalidation) |
| Client state | Zustand (auth session, persisted to `localStorage`) |
| Realtime | native `WebSocket` via a `useRealtime` hook |
| UI primitives | hand-rolled, `class-variance-authority` + `tailwind-merge` (`cn`) |
| Icons | `lucide-react` |

**Design language:** clean, modern, high-contrast. Neutral zinc surfaces, a single
indigo accent, generous rounding (`rounded-xl`), soft shadows, `Geist` type. Every
async surface has explicit loading / empty / error states.

---

## 3. Layers

```
components/ui/*        design-system primitives (Button, Card, Input, Badge, …)
components/shop/*      EventCard, SeatMap, Nav
components/admin/*     Sidebar, StatCard
lib/api/*              one module per gateway surface (auth, catalog, reservation…)
lib/store/auth-store   Zustand session (token + user), persisted
lib/hooks/*            TanStack Query hooks + useRealtime
lib/api/client         fetch wrapper: base URL, Bearer token, X-Admission-Token
```

`lib/api/client` centralizes: the gateway base URL, attaching the access token,
attaching the waiting-room admission token on booking calls, and normalizing
errors. No component calls `fetch` directly.

---

## 4. Key flows

**Discovery → booking (buyer)**
1. `/events` — searchable, filterable grid (genre/location).
2. `/events/[id]` — detail + live **SeatMap**; a `useRealtime('event:{id}')`
   subscription flips seats as `seat.locked/booked/released` arrive.
3. Select seats → `POST /holds` (Bearer + admission token) → redirect to
   `/checkout/[holdId]`.
4. Checkout review → `POST /checkout` (Idempotency-Key) → poll/subscribe order.
5. `/tickets` — issued QR tickets (rendered client-side from the hash).

**Gated events (waiting room)**
`/queue/[eventId]` joins the queue, polls `GET /queue/:id/status`, and on
`admitted` stores the admission token (used by booking calls) and routes on.

**Staff (organizer/admin)**
`/admin` dashboard (analytics), `/admin/events` + `/admin/venues` (create /
publish / price), `/admin/users` (admin-only role management).

---

## 5. Env

```
NEXT_PUBLIC_API_URL   gateway REST base   (default http://localhost:8080)
NEXT_PUBLIC_WS_URL    gateway WS base     (default ws://localhost:8080)
```

The client is otherwise config-free; all routing/auth/rate-limiting lives at the
Gateway.
