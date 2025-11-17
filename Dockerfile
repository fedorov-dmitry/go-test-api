# syntax=docker/dockerfile:1
# --- Build stage ---
FROM golang:1.22-alpine AS builder

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

COPY --from=builder /app/test-api-go /app/test-api-go
COPY scripts/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/test-api-go"]


