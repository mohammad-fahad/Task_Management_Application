# ============================================================================
# Stage 1: Build the Go binary
# ============================================================================
FROM golang:1.22-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/api ./cmd/api

# ============================================================================
# Stage 2: Minimal runtime image
# ============================================================================
FROM alpine:3.19
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 1001 appuser

COPY --from=builder /app/api .
COPY --from=builder /app/migrations ./migrations

USER appuser

EXPOSE 8080

CMD ["./api"]