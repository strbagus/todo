# Build stage
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Cache go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Compile statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -extldflags '-static'" -o /bin/todo ./cmd/server

# Final runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Create directory for persistent SQLite database
RUN mkdir -p /data

COPY --from=builder /bin/todo /app/todo

# Expose HTTP port and volume mount
EXPOSE 8080
VOLUME ["/data"]

ENV PORT=8080 \
    DB_PATH=/data/todo.db

ENTRYPOINT ["/app/todo"]
