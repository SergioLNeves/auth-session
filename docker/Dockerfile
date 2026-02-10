# Stage 1: Build
FROM golang:1.25-bookworm AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate RSA keys for JWT signing
RUN openssl genrsa -out private-key.pem 2048 && \
    chmod 600 private-key.pem && \
    openssl rsa -in private-key.pem -pubout -out public-key.pem

# Build with CGO enabled (required by mattn/go-sqlite3)
RUN CGO_ENABLED=1 go build -o /app/server ./cmd/api/main.go

# Stage 2: Runtime
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates libsqlite3-0 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /build/private-key.pem .
COPY --from=builder /build/public-key.pem .
COPY --from=builder /build/assets ./assets

RUN mkdir -p /data

ENV ENV=production
ENV PORT=8080
ENV LOG_LEVEL=info
ENV PRIVATE_KEY_PATH=/app/private-key.pem
ENV PUBLIC_KEY_PATH=/app/public-key.pem
ENV DB_PATH=/data/auth-session.db

EXPOSE 8080

CMD ["/app/server"]
