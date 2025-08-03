# docker/migrator.Dockerfile
# Multi-stage build for the migration tool
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY backend/ ./backend/

# Build the migrator binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrator ./backend/cmd/migrator

# Runtime image
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata postgresql-client

# Create app user
RUN addgroup -g 1001 app && \
    adduser -u 1001 -G app -s /bin/sh -D app

# Create directories
RUN mkdir -p /app/data /app/logs /app/migrations && \
    chown -R app:app /app

WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=app:app /build/migrator /app/migrator

# Copy migration files
COPY --chown=app:app database/migrations/ ./migrations/

USER app

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./migrator --health-check || exit 1

ENTRYPOINT ["./migrator"]
