# payments

A small Go HTTP service that moves funds between accounts: account creation, JWT login, and an idempotent transfer endpoint. Go 1.25.5, PostgreSQL 17, Redis, Gin, MIT.

## What it guarantees today

- **Atomic transfer.** Debit, credit, and the `transactions` insert run in one SQL transaction with `defer tx.Rollback()` and a single `tx.Commit()`. A failure leaves no partial transfer.
- **No negative balance.** `account.balance` has `CHECK (balance >= 0)`. An insufficient balance aborts the debit and rolls the whole transfer back.
- **Unique `external_id`.** `transactions.external_id` is `UNIQUE NOT NULL`, so a duplicate cannot be committed even if Redis loses the key.
- **A bounded idempotency window.** Redis `SETNX` on `external_id` stores `PROCESSING` for **30 s**; on success the key becomes `COMPLETED` with **10 s** TTL. The deduplication guarantee lasts only as long as that TTL; after it, Postgres is the only guard.

## Quickstart
```bash
cp .env.template .env    # fill in the values below
docker compose up -d
./run-migrations.sh
```

| Variable | Used for |
|---|---|
| `POSTGRES_HOST` | Postgres host (`db` for the Compose API container) |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | Postgres credentials and database |
| `POSTGRES_PORT_EXTERNAL` | Host port for Postgres; also the port migrations connect to |
| `API_PORT` | Port published by Compose for the API. The binary always listens on `:8080`, so use `API_PORT=8080`. |
| `REDIS_PASSWORD` | Redis password |
| `REDIS_PORT` | Host port for Redis |
| `JWT_SECRET_TOKEN` | HS256 signing secret |

```bash
# create an account
curl -s -X POST localhost:8080/api/v1/accounts \
  -H 'Content-Type: application/json' \
  -d '{"user_name":"Ada","user_cpf_cnpj":"12345678901","user_email":"ada@example.com","password":"secret123"}'

# log in; copy access_token into $ACCESS_TOKEN
curl -s -X POST localhost:8080/api/v1/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ada@example.com","password":"secret123"}'

# transfer from the logged-in user's account to another account id
curl -s -X POST localhost:8080/api/v1/transactions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"external_id":"txn-0001","to_account_id":"<destination-account-id>","amount":1000,"type":"TRANSFER","currency":"BRL"}'
```

## API
| Method | Path | Auth | Responses |
|---|---|---|---|
| `POST` | `/api/v1/accounts` | public | `201` created · `400` invalid JSON · `409` email or CPF/CNPJ already registered · `500` internal |
| `POST` | `/api/v1/login` | public | `200` access + refresh token · `400` invalid JSON · `401` invalid credentials |
| `POST` | `/api/v1/transactions` | JWT (`Authorization: Bearer <token>`) | `201` processed · `200` `"transaction already processed"` · `409` `"transaction still being processed"` · `400` invalid JSON · `401` missing/expired token · `500` internal (includes insufficient balance) |

`from_account_id` in the request body is ignored — the origin is the account of the JWT email. Duplicates that already completed, or that hit the database unique constraint, return `200`; `409` is only a request that arrives while the first one is still inside the 30 s `PROCESSING` window.

## Data model
```
account
  id            UUID PK DEFAULT gen_random_uuid()
  balance       BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
  user_name     VARCHAR(120) NOT NULL
  user_cpf_cnpj VARCHAR(14) UNIQUE NOT NULL
  user_email    VARCHAR(120) UNIQUE NOT NULL
  password_hash VARCHAR(255) NOT NULL
  blocked       BOOLEAN DEFAULT FALSE
  created_at    TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
  -- indexes: user_cpf_cnpj, user_email, blocked

transactions
  id              UUID PK DEFAULT gen_random_uuid()
  external_id     VARCHAR(255) UNIQUE NOT NULL
  from_account_id UUID NOT NULL FK -> account(id)
  to_account_id   UUID NOT NULL FK -> account(id)
  type            VARCHAR(50) NOT NULL
  amount          BIGINT NOT NULL CHECK (amount >= 1)
  currency        VARCHAR(10) NOT NULL
  status          VARCHAR(50) NOT NULL
  created_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
  -- indexes: from_account_id, to_account_id, status, created_at DESC, type, (from_account_id, status)
```
Constraints that matter: `balance >= 0`, `amount >= 1`, both foreign keys, and `UNIQUE (external_id)`.

## Architecture

```
client ──► controller (Gin) ──► usecase ──► repository ──► PostgreSQL
                                   │
                                   └──► cache.RedisTransactionStore ──► Redis
```

`cmd/api` wires everything; layers live in `internal/` (`auth`, `cache`, `controllers`, `database/{migrations,postgres,redis}`, `middleware`, `models`, `repository`, `usecases`). Migrations are golang-migrate files under `internal/database/migrations/`, applied by `run-migrations.sh`. The transaction usecase depends on a **concrete** `*cache.RedisTransactionStore`, not an interface, so this is a layered design rather than a port/adapter one. It is not Clean Architecture.

## Decisions

| Decision | Alternative | Why | Cost |
|---|---|---|---|
| Redis `SETNX` for idempotency | A `processed_requests` table in Postgres | One atomic round trip on the hot path, before any SQL work. | Keys expire: after 10 s (`COMPLETED`) the cache no longer knows the transaction; the database must catch it. |
| `balance` as a column | Balance derived from append-only entries | A transfer is two `UPDATE`s plus one insert; reads need no aggregation. | Balance is mutable state; there is no history to replay or audit. |
| JWT HS256 | RS256 | One shared secret, no key distribution; adequate for a single service. | Every verifier can also mint tokens; no per-service key separation or rotation. |

## Known limitations

- No tests anywhere in the repository.
- Idempotency is time-bounded by Redis TTLs (30 s `PROCESSING`, 10 s `COMPLETED`).
- Duplicate detection relies on `strings.Contains(err.Error(), "unique constraint")` — driver text, not a typed error.
- Logging is `log.Println` plus Gin defaults; no metrics, no tracing.
- No reconciliation job comparing transfers with external settlement.
- `balance` is mutated in place. This is not a ledger.

## What I'd do differently in production

- Append-only entries with balance derived or snapshotted.
- Persisted idempotency records keyed by `external_id` plus a payload hash.
- Transactional outbox for downstream effects.
- A versioned OpenAPI contract instead of README payloads.

## License
MIT — see `LICENSE`.
