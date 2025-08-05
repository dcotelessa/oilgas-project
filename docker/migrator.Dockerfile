# docker/migrator.Dockerfile
# Corrected Dockerfile for migration tool with proper Go module paths

FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files from backend directory (where they actually are)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ ./

# Build the migrator binary (from current directory since we copied backend/ to root)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrator ./cmd/migrator

# Runtime image
FROM alpine:3.18

# Install runtime dependencies including golang-migrate
RUN apk add --no-cache ca-certificates tzdata postgresql-client curl && \
    # Install golang-migrate
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Create app user
RUN addgroup -g 1001 app && \
    adduser -u 1001 -G app -s /bin/sh -D app

# Create directories
RUN mkdir -p /app/data /app/logs /app/migrations && \
    chown -R app:app /app

WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=app:app /build/migrator /app/migrator

# Copy migration files (these will be mounted at runtime)
# COPY --chown=app:app database/migrations/ ./migrations/

USER app

# Default environment variables
ENV DEV_DATABASE_URL=""
ENV TEST_DATABASE_URL=""
ENV MIGRATION_PATH="/app/migrations"
ENV LOG_PATH="/app/logs"

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD migrate -path $MIGRATION_PATH -database $DEV_DATABASE_URL version 2>/dev/null || exit 1

# Entry point that can run migrations or the custom migrator
ENTRYPOINT ["/bin/sh", "-c", "\
    if [ \"$1\" = \"schema\" ]; then \
        echo 'Running schema migrations...'; \
        if [ -n \"$DEV_DATABASE_URL\" ]; then \
            echo 'Migrating development database...'; \
            migrate -path $MIGRATION_PATH -database $DEV_DATABASE_URL up; \
        fi; \
        if [ -n \"$TEST_DATABASE_URL\" ]; then \
            echo 'Migrating test database...'; \
            migrate -path $MIGRATION_PATH -database $TEST_DATABASE_URL up; \
        fi; \
    elif [ \"$1\" = \"data\" ]; then \
        echo 'Running data migration...'; \
        ./migrator; \
    else \
        echo 'Running both schema and data migrations...'; \
        if [ -n \"$DEV_DATABASE_URL\" ]; then \
            echo 'Migrating development database schema...'; \
            migrate -path $MIGRATION_PATH -database $DEV_DATABASE_URL up; \
        fi; \
        if [ -n \"$TEST_DATABASE_URL\" ]; then \
            echo 'Migrating test database schema...'; \
            migrate -path $MIGRATION_PATH -database $TEST_DATABASE_URL up; \
        fi; \
        echo 'Running data migration...'; \
        ./migrator; \
    fi", "--"]

# Default command runs both schema and data migrations
CMD ["all"]
