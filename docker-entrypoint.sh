#!/bin/sh
set -e

# Load defaults if not specified
DB_HOST="${DB_HOST:-db}"
DB_PORT="${DB_PORT:-3306}"

echo "Checking if database is available at $DB_HOST:$DB_PORT..."

# Wait loop checking port connectivity
until nc -z -w 2 "$DB_HOST" "$DB_PORT"; do
  echo "Database at $DB_HOST:$DB_PORT is not available yet, waiting..."
  sleep 2
done

echo "Database is ready!"

# Run migrations if enabled
if [ "$RUN_MIGRATIONS" = "true" ]; then
  echo "Starting database migration/seeding..."
  /app/migrate
else
  echo "Skipping migration runner (RUN_MIGRATIONS is not set to true)."
fi

# Execute the main web server process
echo "Starting Kariakoo Inventory web server..."
exec /app/web
