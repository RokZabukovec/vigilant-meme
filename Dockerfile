# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY src/go.mod ./

# Copy go.sum if it exists (optional for standard library only projects)
COPY src/go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY src/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o clip ./cmd/clip

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S clip && \
    adduser -u 1001 -S clip -G clip

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/clip .

# Change ownership to non-root user
RUN chown -R clip:clip /app

# Switch to non-root user
USER clip

# Expose ports
EXPOSE 8080 9999/udp

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/status || exit 1

# Run the application
CMD ["./clip", "-id=clip-node", "-port=8080"]