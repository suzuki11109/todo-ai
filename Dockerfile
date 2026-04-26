# Multi-stage build for production
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Build application
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# Production stage
FROM alpine:latest

# Install PostgreSQL client tools and bash
RUN apk --no-cache add ca-certificates tzdata postgresql-client bash

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy migrations
COPY migrations/ ./migrations/

# Copy static files
COPY static/ ./static/

# Copy entrypoint and migration scripts
COPY scripts/entrypoint.sh /entrypoint.sh
COPY scripts/migrate.sh /app/scripts/migrate.sh
RUN chmod +x /entrypoint.sh /app/scripts/migrate.sh

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

USER appuser

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]