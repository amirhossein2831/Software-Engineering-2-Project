# Shared Infrastructure

Shared services used by every microservice: Postgres, Redis, Kafka (KRaft), Mailhog.
Each microservice lives in its own top-level directory with its own Dockerfile and
is wired into this compose file.

## Start the stack

```bash
cd build/infra
docker compose up -d
docker compose ps
```

## Endpoints

| Service  | Host address        | Notes                                             |
|----------|---------------------|---------------------------------------------------|
| Postgres | `localhost:5432`    | user/pass/db `ticketing`; one DB per service (`auth`, `catalog`, …) |
| Redis    | `localhost:6379`    | seat locks, waiting-room queue                    |
| Kafka    | `localhost:29092`   | host listener; in-network brokers use `kafka:9092`|
| Mailhog  | SMTP `localhost:1025`, UI `http://localhost:8025` | catches outbound email          |

## Tear down

```bash
docker compose down       # keep data
docker compose down -v    # also drop the postgres volume
```
