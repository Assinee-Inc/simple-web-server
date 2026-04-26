# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o simple-web-server ./cmd/web

# Final stage - minimal runtime image
FROM alpine:3.19

WORKDIR /app_src

# Install runtime dependencies for database drivers and email
RUN apk add --no-cache ca-certificates tzdata

# Copy binary to /app/bin (outside the mounted volume)
COPY --from=builder /app/simple-web-server /app/bin/simple-web-server
COPY --from=builder /app/env.template /app/env.template

# Create non-root user for security
RUN adduser -D -u 1000 appuser
USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/bin/simple-web-server"]