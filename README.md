## go-test-api

A small Go service that ingests currency exchange rates into PostgreSQL and exposes HTTP endpoints to query the latest and historical rates.

### Features
- Fetches rates for a configured set of base currencies from a public currency API
- Persists rates in PostgreSQL with an upsert on (date, base, currency)
- HTTP API on port `8080` to retrieve:
  - Latest rate for a given base/currency
  - All rates for a given base on a specific date

## Tech
- Go (see `go.mod` for version)
- PostgreSQL
- pgx v5

## Prerequisites
- Go installed (version per `go.mod`)
- PostgreSQL running and reachable via a connection string

## Database setup
Create the schema/table expected by the service:

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

## Configuration
Configuration is provided via command-line flags:

- `-connection-string` (string): PostgreSQL connection string. Default: `postgresql://postgres:password@localhost:5432/currencies`
- `-api-base-url` (string): Base URL for the currency API. Default: `https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api`
- `-api-key` (string): API key for the currency API. Default: `4a7a2a118a1216b06dc88819dd824eb7` (currently not used by the implementation)
- `-currencies` (string): Comma-separated list of currencies to ingest (e.g., `EUR,USD,RUB,JPY`). Default: `EUR,USD,RUB,JPY`
- `-days-look-back` (int): How many days back to also ingest. Default: `1` (includes today and yesterday)

## Run

```bash
go run ./cmd/test-api-go \
  -connection-string "postgresql://postgres:password@localhost:5432/currencies" \
  -api-base-url "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api" \
  -api-key "your-api-key" \
  -currencies "EUR,USD,RUB,JPY" \
  -days-look-back 1
```

This will:
1. Connect to PostgreSQL
2. Ingest rates for the configured currencies for the configured day range
3. Start the HTTP server on `:8080`

### Build
```bash
go build -o bin/test-api-go ./cmd/test-api-go
```

## API
Base URL: `http://localhost:8080`

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
curl "http://localhost:8080/rates/latest?base=usd&currency=eur"
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
curl "http://localhost:8080/rates/historical?base=usd&date=2025-01-13"
```

## Notes
- Server listens on port `8080` (see `internal/api/server.go`).
- The API serializes Go struct field names as-is (e.g., `Date`, `Base`, `Currency`, `Rate`).
- The external currency API base URL uses a date suffix in the form `@YYYY-MM-DD`.

## License
MIT (or your preferred license)


### Docker

Build and run the service in a container:

```bash
# Build image
docker build -t go-test-api .

# Run (requires a running PostgreSQL; see compose below)
docker run --rm -p 8080:8080 \
  --name go-test-api \
  --env DB_HOST=host.docker.internal \
  --env DB_PORT=5432 \
  go-test-api \
  /app/test-api-go \
  -connection-string "postgresql://postgres:password@host.docker.internal:5432/currencies" \
  -api-base-url "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api" \
  -api-key "your-api-key" \
  -currencies "EUR,USD,RUB,JPY" \
  -days-look-back 1
```

### Docker Compose (API + PostgreSQL)

This repository includes a `docker-compose.yml` that starts PostgreSQL and the API together. The database is initialized automatically with the required schema/table.

```bash
docker compose up --build
# or detached:
docker compose up -d --build
```

Once both services are healthy:
- API: `http://localhost:8080`
- Postgres: `localhost:5432` (user: `postgres`, password: `password`, db: `currencies`)

Example requests:

```bash
curl "http://localhost:8080/rates/latest?base=usd&currency=eur"
curl "http://localhost:8080/rates/historical?base=usd&date=2025-01-13"
```

Stop everything:

```bash
docker compose down -v
```

