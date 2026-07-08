# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency manifests
COPY go.mod ./
# RUN go mod download # Skipped per rule: do not install dependencies yet

# Copy source code
COPY . .

# Build the application binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o worker ./cmd/worker/main.go

# Production stage
FROM alpine:3.19 AS runner

WORKDIR /app

# Copy binaries from build stage
COPY --from=builder /app/api .
COPY --from=builder /app/worker .
# Copy migration files for runtime migrations
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./api"]
