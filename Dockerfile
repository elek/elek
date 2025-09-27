# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY main.go .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o smokeping .

# Runtime stage
FROM alpine:latest

# Install mtr and ca-certificates
RUN apk --no-cache add mtr ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh smokeping

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/smokeping .

# Change ownership to non-root user
RUN chown smokeping:smokeping /app/smokeping

# Switch to non-root user
USER smokeping

# Expose default port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
ENTRYPOINT ["./smokeping"]