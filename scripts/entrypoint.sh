#!/bin/sh
set -e

# Wait for database to be ready
echo "Waiting for database..."
while ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"; do
  sleep 1
done

echo "Database is ready!"

# Run migrations
echo "Running database migrations..."
if [ -f "/app/scripts/migrate.sh" ]; then
    /app/scripts/migrate.sh up
else
    echo "Migration script not found!"
    exit 1
fi

echo "Migrations completed!"

# Start the application
if command -v air >/dev/null 2>&1 && [ -f ".air.toml" ]; then
    echo "Starting with Air hot reload..."
    exec air -c .air.toml
else
    echo "Starting server..."
    exec /app/server
fi