# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency manifests
COPY go.mod ./
# RUN go mod download # Skipped per rule: do not install dependencies yet

# Copy source code
COPY . .

# Build the application binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Production stage
FROM alpine:3.19 AS runner

WORKDIR /app

# Copy binary from build stage
COPY --from=builder /app/main .
# Copy migration files for runtime migrations
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./main"]
