# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway main.go

# Stage 2: Create a small final image
FROM alpine:3.19

# Install CA certificates (important for HTTPS/TLS)
RUN apk add --no-cache ca-certificates

# Set working directory in the final container
WORKDIR /app

# Copy the compiled Go binary from builder stage
COPY --from=builder /app/gateway .

# Expose port for WebSocket server
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./gateway"]