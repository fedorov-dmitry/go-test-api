## go-test-api

A small Go service that ingests currency exchange rates into PostgreSQL and exposes HTTP endpoints to query the latest and historical rates.

### Features
- Fetches rates for a configured set of base currencies from a public currency API
- Persists rates in PostgreSQL with an upsert on (date, base, currency)
- HTTP API on port `8088` to retrieve:
  - Latest rate for a given base/currency
  - All rates for a given base on a specific date

## Tech
- Go (see `go.mod` for version)
- PostgreSQL
- pgx v5
- Docker, Docker Compose (optional)

## Prerequisites (when running locally without Docker)
- Go installed (version per `go.mod`)
- PostgreSQL reachable (or run with Docker Compose)

## Database setup
If you are NOT using Docker Compose, create the schema/table expected by the service (or run `db/init.sql`):

```sql
CREATE SCHEMA IF NOT EXISTS app;

CREATE TABLE IF NOT EXISTS app.currency_rates (
  date      date    NOT NULL,
  base      text    NOT NULL,
  currency  text    NOT NULL,
  rate      double precision NOT NULL,
  PRIMARY KEY (date, base, currency)
);
```

## Configuration (environment variables)
- `APP_PORT`: HTTP port the server listens on. Default: `8088`
- `CONNECTION_STRING`: PostgreSQL connection string. Default in containers: `postgresql://postgres:password@postgres:5432/currencies` (use `localhost` when running locally)
- `API_BASE_URL`: Currency API base URL. Default: `https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api`
- `CURRENCIES`: Comma-separated list of currencies to ingest, e.g. `EUR,USD,RUB,JPY`. Default: `EUR,USD,RUB,JPY`
- `DAYS_LOOK_BACK`: Non-negative integer; number of days back to ingest in addition to today. Default: `1`
  - Note: The service ingests for days in range `[0..DAYS_LOOK_BACK]` (inclusive). For example, `1` means today and yesterday.
- Container-only helpers (used by entrypoint wait logic):
  - `DB_HOST` (default: `postgres`)
  - `DB_PORT` (default: `5432`)
  - `WAIT_FOR_DB` (default: `true`)

## Run

Local run (without Docker), with env vars:

```bash
export APP_PORT=8088
export CONNECTION_STRING="postgresql://postgres:password@localhost:5432/currencies"
export API_BASE_URL="https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api"
export CURRENCIES="EUR,USD,RUB,JPY"
export DAYS_LOOK_BACK=1

go run ./cmd/test-api-go
```

This will:
1. Connect to PostgreSQL
2. Ingest rates for the configured currencies for the configured day range
3. Start the HTTP server on `:8088`

### Build
```bash
go build -o bin/test-api-go ./cmd/test-api-go
# run the built binary (env vars as above)
./bin/test-api-go
```

## API
Base URL: `http://localhost:8088`

### GET `/rates/latest`
- Query params: `base` (string, required), `currency` (string, required)
- Returns a single object:
  ```json
  {
    "Date": "2025-01-14T00:00:00Z",
    "Base": "usd",
    "Currency": "eur",
    "Rate": 0.92
  }
  ```

Example:
```bash
curl "http://localhost:8088/rates/latest?base=usd&currency=eur"
```

### GET `/rates/historical`
- Query params: `base` (string, required), `date` (YYYY-MM-DD, required)
- Returns an array of objects for all stored target currencies for that base on the given date:
  ```json
  [
    { "Date": "2025-01-13T00:00:00Z", "Base": "usd", "Currency": "eur", "Rate": 0.92 },
    { "Date": "2025-01-13T00:00:00Z", "Base": "usd", "Currency": "jpy", "Rate": 145.1 }
  ]
  ```

Example:
```bash
curl "http://localhost:8088/rates/historical?base=usd&date=2025-01-13"
```

## Notes
- Server listens on `APP_PORT` (default `8088`, see `internal/api/server.go`).
- The API serializes Go struct field names as-is (e.g., `Date`, `Base`, `Currency`, `Rate`).
- `base` and `currency` values are normalized to lowercase internally.
- The external currency API base URL uses a date suffix in the form `@YYYY-MM-DD`.

## License
MIT (or your preferred license)


### Docker

Build and run the service in a container:

```bash
# Build image
docker build -t go-test-api .

# Run (requires a running PostgreSQL; see Compose below). On macOS/Windows, use host.docker.internal
docker run --rm -p 8088:8088 \
  --name go-test-api \
  -e APP_PORT=8088 \
  -e CONNECTION_STRING="postgresql://postgres:password@host.docker.internal:5432/currencies" \
  -e API_BASE_URL="https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api" \
  -e CURRENCIES="EUR,USD,RUB,JPY" \
  -e DAYS_LOOK_BACK=1 \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=5432 \
  go-test-api
```

### Docker Compose (API + PostgreSQL)

This repository includes a `docker-compose.yml` that starts PostgreSQL and the API together. The database is initialized automatically with the required schema/table.

```bash
docker compose up --build
# or detached:
docker compose up -d --build
```

Once both services are healthy:
- API: `http://localhost:8088`
- Postgres: `localhost:5432` (user: `postgres`, password: `password`, db: `currencies`)

Example requests:

```bash
curl "http://localhost:8088/rates/latest?base=usd&currency=eur"
curl "http://localhost:8088/rates/historical?base=usd&date=2025-01-13"
```

Stop everything:

```bash
docker compose down -v
```

