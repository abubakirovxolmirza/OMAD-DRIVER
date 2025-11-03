# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o taxi-service cmd/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/taxi-service .

# Copy .env.example (user should provide actual .env)
COPY --from=builder /app/.env.example .

# Create upload directory
RUN mkdir -p uploads/avatars uploads/licenses

# Expose port
EXPOSE 8080

# Run application
CMD ["./taxi-service"]
