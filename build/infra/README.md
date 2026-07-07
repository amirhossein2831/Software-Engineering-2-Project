# Application Stack

Runs the whole application: shared infra (Postgres, Redis, Kafka KRaft, Mailhog),
every microservice, the gateway, the Next.js frontend, and the NGINX edge. Each
microservice lives in its own top-level directory with its own Dockerfile and is
wired into this compose file. Helper scripts live in `build/scripts/`, and shared
config (`.env.example`, `nginx.conf`) lives in `build/config/`.

> The self-hosted **GitLab CI** stack is separate — see `build/gitlab/`. It runs
> on its own network and is not part of this compose file.

## Start the stack

```bash
cd build/infra
cp ../config/.env.example .env   # optional — safe dev defaults exist
docker compose up -d --build
docker compose ps
```

## Endpoints

| Service  | Host address        | Notes                                             |
|----------|---------------------|---------------------------------------------------|
| Edge     | `http://localhost:8090` | NGINX — web client + `/api` → gateway (`EDGE_PORT`) |
| Postgres | `localhost:5432`    | user/pass/db `ticketing`; one DB per service (`auth`, `catalog`, …) |
| Redis    | `localhost:6379`    | seat locks, waiting-room queue                    |
| Kafka    | `localhost:29092`   | host listener; in-network brokers use `kafka:9092`|
| Mailhog  | SMTP `localhost:1025`, UI `http://localhost:8025` | catches outbound email          |
| Auth     | `http://localhost:8081`                           | register / login / JWT          |

## Tear down

```bash
docker compose down       # keep data
docker compose down -v    # also drop the postgres volume
```
