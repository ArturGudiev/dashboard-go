# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for better layer caching (replace ./internal/go-fuzzyfinder must exist)
COPY go.mod go.sum ./
COPY internal/go-fuzzyfinder ./internal/go-fuzzyfinder
RUN go mod download

# Copy source
COPY . .

# Build binary (-buildvcs=false: no git in build context)
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /dashboard .

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /dashboard .

# Optional: copy .env into image (prefer runtime env: docker run --env-file .env ...)
# COPY --from=builder /app/.env .

EXPOSE 8080

ENTRYPOINT ["./dashboard"]
