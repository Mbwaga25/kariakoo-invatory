# Build Stage
FROM golang:alpine AS builder

WORKDIR /app

# Install git and other build essentials if needed
RUN apk add --no-cache git

# Copy dependency manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the web server and migration tool binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o web cmd/web/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o migrate cmd/migrate/main.go

# Production Stage
FROM alpine:3.20

WORKDIR /app

# Install netcat-openbsd for checking DB connection in the entrypoint script
RUN apk add --no-cache netcat-openbsd

# Copy compiled binaries from the builder stage
COPY --from=builder /app/web /app/web
COPY --from=builder /app/migrate /app/migrate

# Copy static assets and HTML templates
COPY --from=builder /app/ui /app/ui

# Copy migration SQL and seed directories
COPY --from=builder /app/cmd/migrate/sql /app/cmd/migrate/sql
COPY --from=builder /app/cmd/migrate/seed /app/cmd/migrate/seed

# Copy entrypoint script
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Expose the application port
EXPOSE 8081

ENTRYPOINT ["/app/docker-entrypoint.sh"]
