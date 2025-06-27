# Stage 1: Build the Go application
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# -ldflags="-s -w" strips debug info to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/myapp

# Stage 2: Create minimal runtime image
FROM alpine:3.19

WORKDIR /app

# Copy built binary from builder
COPY --from=builder /app/main .

# Create a non-root user (good practice)
RUN addgroup -S mygroup && \
    adduser -S myuser -G mygroup && \
    chown -R myuser:mygroup /app
USER myuser

# Application port (adjust as needed)
EXPOSE 3456
EXPOSE 3457

# Run the binary
CMD ["./main"]

