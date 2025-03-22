FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY src/rivian/go.mod src/rivian/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY src/rivian/ .

# Build both binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o auth cmd/auth/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o logger cmd/logger/main.go

# Create final image
FROM alpine:latest

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/auth .
COPY --from=builder /app/logger .

# Create directory for auth file
RUN mkdir -p /app && chmod 700 /app

# Set default command (can be overridden by docker-compose)
ENTRYPOINT ["./logger"] 