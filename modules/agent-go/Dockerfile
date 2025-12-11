# Dev stage with hot-reload
FROM golang:1.24-alpine AS dev
WORKDIR /app
RUN apk add --no-cache git ca-certificates
RUN go install github.com/air-verse/air@v1.52.3
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /agent-go ./cmd/agent
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /etl ./cmd/etl

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

# Copy binaries
COPY --from=builder /agent-go /agent-go
COPY --from=builder /migrate /migrate
COPY --from=builder /etl /etl

# Copy migrations and seeders
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/seeders /app/seeders

# Set ownership
RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080
CMD ["/agent-go"]
