# Event Ticketing Platform

A distributed event-ticketing system: event catalog, concurrency-safe seat
booking, an orchestrated checkout saga, a virtual waiting room, and real-time
seat/order updates. Polyglot monorepo — each service is a self-contained
top-level directory; a Next.js web client sits behind an NGINX edge.

Architecture: [`docs/architecture/`](docs/architecture/).

## Services

| Service | Port | Role |
|---|---|---|
| nginx | 80 | single edge — serves the web client, proxies `/api` → gateway |
| frontend | 3000 | Next.js web client (buyer + organizer/admin) |
| gateway | 8080 | API edge: JWT verify, rate limit, admission check, routing |
| auth | 8081 | register/login/JWT/roles; publishes `user.registered` |
| catalog | 8082 | events, venues, sectors, pricing, seat-map layout |
| reservation | 8083 | live seat state, Redis seat locks, holds, expiry reaper |
| checkout | 8084 | order lifecycle + checkout saga (charge → commit / compensate) |
| ticketing | 8085 | consumes `payment.succeeded` → issues QR tickets |
| notification | 8086 | outbox + worker → email (Mailhog) / SMS on events |
| realtime | 8087 | WebSocket fan-out hub off the Kafka stream |
| waiting-room | 8088 | Redis queue + signed admission tokens |
| postgres / redis / kafka / mailhog | — | infra (database-per-service, locks/queue, event bus, mail) |

## Run the whole stack

```bash
# 1. Generate go.sum for every service (Docker builds need it):
bash build/scripts/tidy-all.sh

# 2. Configure (optional — safe dev defaults exist):
cp build/config/.env.example build/infra/.env

# 3. Boot everything:
docker compose -f build/infra/docker-compose.yml up --build
```

Then open **http://localhost:8090** (the web client) and
**http://localhost:8025** (Mailhog, to see sent emails).

> **Ports & isolation.** This compose runs as its own project (`ticketing-app`),
> so it never touches an unrelated stack — you can keep GitLab (or anything else)
> running. The edge defaults to **8090** to stay clear of GitLab on 80; override
> with `EDGE_PORT`. Mailhog's UI port is `MAILHOG_PORT` (default 8025).
> The single NGINX edge means the browser uses one origin — the client calls a
> relative `/api`, and WebSocket upgrades are proxied through `/api/ws`.

## Notes

- `build/infra/docker-compose.yml` runs the whole application (infra + all
  services + edge). Config lives in `build/config/` (`.env.example`, `nginx.conf`).
- `build/gitlab/docker-compose.yml` is the separate, self-hosted **GitLab CI**
  stack — it runs on its own `gitlab-net` network, independent of the app.
- Analytics (the graded bonus — organizer dashboards from Kafka projections) is
  designed but not yet implemented; the admin dashboard has a placeholder.
