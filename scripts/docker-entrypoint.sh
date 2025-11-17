#!/usr/bin/env sh
set -e

# Wait for PostgreSQL if DB_HOST/DB_PORT are provided
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
WAIT_FOR_DB="${WAIT_FOR_DB:-true}"

if [ "${WAIT_FOR_DB}" = "true" ]; then
  echo "Waiting for PostgreSQL at ${DB_HOST}:${DB_PORT}..."
  for i in $(seq 1 60); do
    if nc -z "${DB_HOST}" "${DB_PORT}" >/dev/null 2>&1; then
      echo "PostgreSQL is up."
      break
    fi
    echo "PostgreSQL not ready yet... (${i}/60)"
    sleep 1
  done
fi

echo "Starting: $*"
exec "$@"
