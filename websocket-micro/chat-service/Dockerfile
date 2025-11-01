# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source
COPY . .

# Build the app
RUN CGO_ENABLED=0 GOOS=linux go build -o chat-service main.go

# Stage 2: Lightweight final image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/chat-service .

ENTRYPOINT ["./chat-service"]