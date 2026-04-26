#!/bin/bash

set -e

# Get database connection details from environment or .env file
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-todo_user}
DB_PASSWORD=${DB_PASSWORD:-todo_pass}
DB_NAME=${DB_NAME:-todo_db}

# Use postgres superuser for migrations to create tables
PSQL="PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME}"

run_migration() {
    local direction=$1
    local version=$2

    if [ "$direction" = "up" ]; then
        echo "Running UP migrations..."
        for file in migrations/up/*.up.sql; do
            if [ -f "$file" ]; then
                echo "Applying: $file"
                $PSQL -f "$file" || { echo "Migration failed: $file"; exit 1; }
            fi
        done
    elif [ "$direction" = "down" ]; then
        echo "Running DOWN migrations..."
        # Reverse order for down migrations
        for file in $(ls migrations/down/*.down.sql | sort -r); do
            if [ -f "$file" ]; then
                echo "Reverting: $file"
                $PSQL -f "$file" || { echo "Migration failed: $file"; exit 1; }
            fi
        done
    else
        echo "Invalid direction: $direction"
        exit 1
    fi
}

case "$1" in
    up)
        run_migration "up"
        ;;
    down)
        run_migration "down"
        ;;
    reset)
        run_migration "down"
        run_migration "up"
        ;;
    *)
        echo "Usage: $0 {up|down|reset}"
        exit 1
        ;;
esac

echo "Migration completed successfully!"