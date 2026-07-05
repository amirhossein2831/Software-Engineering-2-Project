# Increment 1 — Monorepo Foundation + IAM Service — Implementation Plan

> **For agentic workers:** Implement task-by-task. Steps use checkbox (`- [ ]`) syntax. TDD: test first, watch it fail, minimal code, watch it pass, commit.

**Goal:** Stand up the Go monorepo foundation (shared packages + local infra via Docker Compose) and deliver a fully working, tested **Identity & Access (IAM)** service: register, login, JWT refresh, logout.

**Architecture:** Single Go module (`ticketing`) rooted at `platform/`, following the **golang-standards/project-layout** convention: shared libraries in `pkg/`, per-service private code in `internal/<service>/`, executables in `cmd/<service>/`, protobuf/API defs in `api/`, container packaging in `build/package/<service>/`, orchestration in `deployments/`, helper scripts in `scripts/`, and the frontend in `web/`. Direct GORM domain=DB mapping per the design. IAM exposes REST via Fiber v3 and persists to its own Postgres database. Infra (Postgres, Redis, Kafka KRaft, Mailhog) runs from one `deployments/docker-compose.yml`.

**Repository layout (golang-standards/project-layout):**
```
platform/
  cmd/<service>/main.go        # executables
  internal/<service>/...       # private per-service code (model, repository, service, handler)
  pkg/...                      # shared libs (config, logger, database, jwtauth)
  api/proto/                   # gRPC/protobuf definitions
  build/package/<service>/     # Dockerfiles
  deployments/                 # docker-compose, k8s manifests, postgres init
  scripts/                     # smoke tests, helpers
  web/                         # Next.js frontend
```

**Tech Stack:** Go 1.26, Fiber v3 (`github.com/gofiber/fiber/v3`), GORM (`gorm.io/gorm` + `gorm.io/driver/postgres`), `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`, Docker Compose.

## Global Constraints

- Module path: `ticketing`. Go version floor: **1.26**.
- Web framework: **Fiber v3** only. ORM: **GORM** only. No separate domain/persistence model split — GORM structs are the domain models.
- Database-per-service: IAM uses database `iam` on the shared Postgres instance.
- Every service reads config from **env vars** (12-factor); no hardcoded secrets.
- Existing `test-service-go/` (nested module) and `build/` GitLab infra are left untouched.
- Tests: pure logic unit-tested without a DB; DB-touching tests use the real Postgres and **skip** when `TEST_DATABASE_URL` is unset (so `go test ./...` is green with no infra).

---

### Task 1: Go module + shared config package

**Files:**
- Create: `go.mod`
- Create: `pkg/config/config.go`
- Test: `pkg/config/config_test.go`

**Interfaces produced:**
- `config.Get(key, fallback string) string`
- `config.MustGet(key string) string` (panics if missing)
- `config.GetInt(key string, fallback int) int`

- [ ] Step 1: `go mod init ticketing`
- [ ] Step 2: Write `config_test.go` — `Get` returns fallback when unset, env value when set; `GetInt` parses; `MustGet` panics when missing.
- [ ] Step 3: Run tests → FAIL (package missing).
- [ ] Step 4: Implement `config.go` using `os.LookupEnv` + `strconv.Atoi`.
- [ ] Step 5: Run tests → PASS.
- [ ] Step 6: Commit `feat(pkg): add env config loader`.

### Task 2: Shared structured logger

**Files:**
- Create: `pkg/logger/logger.go`
- Test: `pkg/logger/logger_test.go`

**Interfaces produced:**
- `logger.New(service string) *slog.Logger` (JSON handler, includes `service` attr, level from `LOG_LEVEL`).

- [ ] Step 1: Test — `New("iam")` returns non-nil logger that emits JSON containing `"service":"iam"` (capture via a `bytes.Buffer` handler variant or assert level parsing).
- [ ] Step 2: Run → FAIL.
- [ ] Step 3: Implement with `slog.NewJSONHandler`.
- [ ] Step 4: Run → PASS.
- [ ] Step 5: Commit `feat(pkg): add structured logger`.

### Task 3: Shared GORM/Postgres connector

**Files:**
- Create: `pkg/database/database.go`
- Test: `pkg/database/database_test.go`

**Interfaces produced:**
- `database.Open(dsn string) (*gorm.DB, error)`
- `database.MustOpen(dsn string) *gorm.DB`

- [ ] Step 1: Test — `Open("")` returns an error; when `TEST_DATABASE_URL` is set, `Open` succeeds and `db.Exec("SELECT 1")` works (skip otherwise).
- [ ] Step 2: Run → FAIL.
- [ ] Step 3: Implement using `gorm.io/driver/postgres` + `gorm.Open`.
- [ ] Step 4: Run → PASS/SKIP.
- [ ] Step 5: Commit `feat(pkg): add gorm postgres connector`.

### Task 4: Local infra via Docker Compose

**Files:**
- Create: `deployments/docker-compose.yml`
- Create: `deployments/.env.example`
- Create: `deployments/README.md`

Infra services: `postgres` (16, multiple DBs via init script), `redis` (7), `kafka` (KRaft, single node), `mailhog`. Add an init script `deployments/postgres/init-databases.sh` creating `iam`, `catalog`, `reservation`, `checkout`, `ticketing`, `notification`, `analytics` databases.

- [ ] Step 1: Write `docker-compose.yml` with the four infra services + healthchecks + named volumes + `ci-net`-style bridge network.
- [ ] Step 2: Write `postgres/init-databases.sh` (loops over DB names, `createdb`).
- [ ] Step 3: `docker compose -f deployments/docker-compose.yml up -d postgres redis kafka mailhog` → verify all healthy (`docker compose ps`).
- [ ] Step 4: Commit `chore(deploy): add local infra compose (postgres/redis/kafka/mailhog)`.

### Task 5: IAM domain models

**Files:**
- Create: `internal/iam/model/user.go`
- Create: `internal/iam/model/refresh_token.go`
- Test: `internal/iam/model/user_test.go`

**Interfaces produced:**
- `model.User{ ID uuid; Email string; PasswordHash string; Role Role; CreatedAt }` with `Role` enum (`buyer|organizer|admin`) and `TableName() "users"`.
- `model.RefreshToken{ ID uuid; UserID uuid; TokenHash string; ExpiresAt; RevokedAt *time }`.
- `model.Role` with `Valid() bool`.

- [ ] Step 1: Test — `Role("buyer").Valid()` true; `Role("x").Valid()` false; default role constant is `RoleBuyer`.
- [ ] Step 2: Run → FAIL.
- [ ] Step 3: Implement models (UUID via `github.com/google/uuid`, gorm tags).
- [ ] Step 4: Run → PASS.
- [ ] Step 5: Commit `feat(iam): add user and refresh-token models`.

### Task 6: Password hashing

**Files:**
- Create: `internal/iam/auth/password.go`
- Test: `internal/iam/auth/password_test.go`

**Interfaces produced:**
- `auth.HashPassword(plain string) (string, error)`
- `auth.CheckPassword(hash, plain string) bool`

- [ ] Step 1: Test — hash of "secret" verifies with CheckPassword; wrong password fails; hash != plain.
- [ ] Step 2: Run → FAIL.
- [ ] Step 3: Implement with `bcrypt.GenerateFromPassword` / `CompareHashAndPassword`.
- [ ] Step 4: Run → PASS.
- [ ] Step 5: Commit `feat(iam): add bcrypt password hashing`.

### Task 7: JWT token service (shared)

**Files:**
- Create: `pkg/jwtauth/jwtauth.go`
- Test: `pkg/jwtauth/jwtauth_test.go`

**Interfaces produced:**
- `jwtauth.NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager`
- `(*Manager) IssueAccess(userID, role string) (string, error)`
- `(*Manager) Verify(token string) (*Claims, error)` where `Claims{ UserID, Role string; jwt.RegisteredClaims }`
- `(*Manager) NewRefreshToken() (raw string, hash string)` (opaque random + sha256 hash)

- [ ] Step 1: Test — issued access token verifies and round-trips `userID`/`role`; tampered token fails; expired token fails; refresh raw≠hash and hash is stable.
- [ ] Step 2: Run → FAIL.
- [ ] Step 3: Implement with `golang-jwt/jwt/v5` (HS256) + `crypto/rand` + `crypto/sha256`.
- [ ] Step 4: Run → PASS.
- [ ] Step 5: Commit `feat(pkg): add jwt manager (access + opaque refresh)`.

### Task 8: IAM repository

**Files:**
- Create: `internal/iam/repository/user_repo.go`
- Test: `internal/iam/repository/user_repo_test.go`

**Interfaces produced:**
- `repository.NewUserRepo(db *gorm.DB) *UserRepo`
- `(*UserRepo) Create(ctx, *model.User) error` (maps unique-violation → `repository.ErrEmailTaken`)
- `(*UserRepo) FindByEmail(ctx, email) (*model.User, error)` (`ErrNotFound`)
- `(*UserRepo) SaveRefresh(ctx, *model.RefreshToken) error`
- `(*UserRepo) FindRefreshByHash(ctx, hash) (*model.RefreshToken, error)`
- `(*UserRepo) RevokeRefresh(ctx, id) error`

- [ ] Step 1: Test (guarded by `TEST_DATABASE_URL`, AutoMigrate in setup, unique tests) — Create+FindByEmail round-trip; duplicate email → `ErrEmailTaken`; refresh save/find/revoke.
- [ ] Step 2: Run → FAIL/SKIP.
- [ ] Step 3: Implement with GORM; detect Postgres unique violation (`errors.Is(err, gorm.ErrDuplicatedKey)`).
- [ ] Step 4: Run → PASS (with infra up + `TEST_DATABASE_URL`).
- [ ] Step 5: Commit `feat(iam): add user repository`.

### Task 9: IAM auth service (use cases)

**Files:**
- Create: `internal/iam/service/auth_service.go`
- Test: `internal/iam/service/auth_service_test.go`

**Interfaces produced:**
- `service.NewAuthService(repo, jwt *jwtauth.Manager, refreshTTL) *AuthService`
- `Register(ctx, email, password) (*model.User, error)`
- `Login(ctx, email, password) (access, refresh string, err error)` (`ErrInvalidCredentials`)
- `Refresh(ctx, rawRefresh) (access, refresh string, err error)` (rotate + revoke old)
- `Logout(ctx, rawRefresh) error`

Test uses a fake repo (in-memory) implementing a `UserStore` interface the service depends on — keeps this test DB-free.

- [ ] Step 1: Extract a `UserStore` interface in the service package matching the repo methods; write fake in test.
- [ ] Step 2: Test — Register hashes password + defaults role buyer; duplicate → error; Login wrong password → `ErrInvalidCredentials`; Login ok → non-empty tokens; Refresh rotates (old hash revoked, new issued); Logout revokes.
- [ ] Step 3: Run → FAIL.
- [ ] Step 4: Implement service.
- [ ] Step 5: Run → PASS.
- [ ] Step 6: Commit `feat(iam): add auth service (register/login/refresh/logout)`.

### Task 10: IAM HTTP handlers (Fiber v3)

**Files:**
- Create: `internal/iam/handler/auth_handler.go`
- Create: `internal/iam/handler/router.go`
- Test: `internal/iam/handler/auth_handler_test.go`

**Interfaces produced:**
- `handler.NewRouter(app *fiber.App, svc *service.AuthService)` registering:
  `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`, `GET /healthz`.
- JSON request/response DTOs; validation returns `400`; auth failures `401`; duplicate `409`.

Test uses `app.Test(httptest.NewRequest(...))` (Fiber's built-in test harness) with a fake service.

- [ ] Step 1: Test — register returns 201 + user id; bad body 400; login returns 200 + tokens; wrong creds 401; healthz 200.
- [ ] Step 2: Run → FAIL.
- [ ] Step 3: Implement handlers + router + DTOs + error mapping.
- [ ] Step 4: Run → PASS.
- [ ] Step 5: Commit `feat(iam): add fiber v3 auth handlers`.

### Task 11: IAM main + Dockerfile + compose wiring

**Files:**
- Create: `cmd/iam/main.go`
- Create: `build/package/iam/Dockerfile`
- Modify: `deployments/docker-compose.yml` (add `iam` service)

- [ ] Step 1: `main.go` — load config, open DB, `AutoMigrate(User, RefreshToken)`, build jwt manager/repo/service/router, `app.Listen(:PORT)`.
- [ ] Step 2: Multi-stage `Dockerfile` (build static binary → distroless/alpine).
- [ ] Step 3: Add `iam` service to compose (depends_on postgres healthy, env `DATABASE_URL`, `JWT_SECRET`, `PORT=8081`, port map `8081:8081`).
- [ ] Step 4: `docker compose up -d --build iam` → `curl :8081/healthz` returns 200.
- [ ] Step 5: Commit `feat(iam): add entrypoint, dockerfile, compose service`.

### Task 12: End-to-end smoke check

**Files:**
- Create: `internal/iam/README.md` (run + curl examples)
- Create: `scripts/smoke-iam.sh` (register → login → refresh → logout via curl, assert status codes)

- [ ] Step 1: Write `smoke-iam.sh` (set -e; curl each endpoint; grep expected status).
- [ ] Step 2: Run against the compose stack → all steps pass.
- [ ] Step 3: Write README with commands.
- [ ] Step 4: Commit `docs(iam): add readme + smoke script`.

---

## Self-Review

- **Spec coverage (this increment):** foundation (module, config, logger, db, infra) ✓; IAM auth + JWT + roles ✓ (maps to design §4.2 IAM). Other services are later increments (own plans).
- **Placeholders:** none — each task has concrete files, interfaces, and commands.
- **Type consistency:** `jwtauth.Manager`, `model.User/Role`, `repository.UserRepo`, `service.AuthService`, `handler.NewRouter` names are used consistently across tasks 5→11.
- **Deferred by design:** gRPC `VerifyToken` server (needed by the Gateway) lands in the Gateway increment; IAM ships REST first.
