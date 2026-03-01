# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git and certificates
RUN apk add --no-cache git ca-certificates

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application statically
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o api ./cmd/api

# Stage 2: Create a minimal production image
FROM alpine:latest

# Install CA certificates to enable HTTPS calls out
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy the binary from the builder stage
COPY --from=builder /app/api .

# Copy the .env file if it exists (for standalone testing, though docker-compose often mounts it or uses env_file)
# We don't strictly require it, config.Load handles missing .env gracefully.
COPY --chown=appuser:appgroup .env* ./

# Change ownership of the binary and working directory
RUN chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Expose the API port
EXPOSE 8080

# Run the binary
CMD ["./api"]
