# Web build stage
FROM node:20-alpine AS web-builder

WORKDIR /web

# Copy web project files
COPY web/package*.json ./

# Install dependencies
RUN npm ci

# Copy web source
COPY web/ ./

# Build web UI
RUN npm run build

# Go build stage
FROM golang:1.25-alpine AS go-builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o task-scheduler ./cmd/scheduler

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS and wget for health checks
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1000 scheduler && \
    adduser -D -u 1000 -G scheduler scheduler

WORKDIR /app

# Copy binary from go-builder
COPY --from=go-builder /app/task-scheduler .

# Copy web-dist from web-builder
COPY --from=web-builder /web-dist ./web-dist

# Copy config example
COPY config.yaml.example ./config.yaml.example

# Change ownership
RUN chown -R scheduler:scheduler /app

# Switch to non-root user
USER scheduler

# Expose ports
EXPOSE 8080 9090 9091

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["./task-scheduler"]
CMD ["-config", "/app/config.yaml"]
