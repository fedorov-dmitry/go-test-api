# syntax=docker/dockerfile:1
# --- Build stage ---
FROM golang:1.25-alpine AS builder

WORKDIR /src
ENV CGO_ENABLED=0 \
    GO111MODULE=on

# Install git for private module fetching if needed
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# Build the binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w" -o /app/test-api-go ./cmd/test-api-go

# --- Runtime stage ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata netcat-openbsd
WORKDIR /app

ENV APP_PORT=8088
ENV DAYS_LOOK_BACK=1
ENV CONNECTION_STRING=postgresql://postgres:password@postgres:5432/currencies
ENV API_BASE_URL=https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api
ENV CURRENCIES=EUR,USD,RUB,JPY
ENV JOB_CRON="* * * * *"
# Defaults for entrypoint database wait helper
ENV DB_HOST=postgres
ENV DB_PORT=5432
ENV WAIT_FOR_DB=true

COPY --from=builder /app/test-api-go /app/test-api-go
COPY scripts/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

EXPOSE 8088
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/test-api-go"]

