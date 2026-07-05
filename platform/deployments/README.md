# Local Infrastructure

One-command local stack for the ticketing platform.

## Start infra

```bash
cd platform/deployments
docker compose up -d            # postgres, redis, kafka (KRaft), mailhog
docker compose ps               # verify all healthy
```

## Endpoints

| Service  | Host address        | Notes                                   |
|----------|---------------------|-----------------------------------------|
| Postgres | `localhost:5432`    | user/pass/db `ticketing`; one DB per service (`iam`, `catalog`, …) |
| Redis    | `localhost:6379`    | seat locks, waiting-room queue          |
| Kafka    | `localhost:29092`   | host listener (`PLAINTEXT_HOST`); in-network brokers use `kafka:9092` |
| Mailhog  | SMTP `localhost:1025`, UI `http://localhost:8025` | catches all outbound email |

## Connection strings

```bash
# from the host (tests): connect to the per-service database, e.g. iam
export TEST_DATABASE_URL="postgres://ticketing:ticketing@localhost:5432/iam?sslmode=disable"
```

## Tear down

```bash
docker compose down             # keep data
docker compose down -v          # also drop the postgres volume
```
