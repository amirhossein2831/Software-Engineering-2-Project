# Ticketing Platform — Web Client

Next.js 15 (App Router) + TypeScript + Tailwind v4 client for the event ticketing
platform. Talks only to the API Gateway (`:8080`) over REST + WebSocket.

Design: [`../docs/architecture/04-client.md`](../docs/architecture/04-client.md).

## Getting started

```bash
cp .env.example .env.local
npm install
npm run dev
```

Open http://localhost:3000.

## Layout

- `src/app/(shop)` — buyer area (browse, seat selection, checkout, tickets)
- `src/app/admin` — organizer + admin area (dashboard, events, venues, users)
- `src/lib/api` — one module per gateway surface
- `src/lib/store` — Zustand auth session
- `src/lib/hooks` — TanStack Query hooks + realtime WebSocket
- `src/components/ui` — design-system primitives
