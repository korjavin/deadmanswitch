FROM golang:1.24-alpine AS builder

# Install Git for dependency downloads
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with CGO disabled
RUN CGO_ENABLED=0 GOOS=linux go build -a -o deadmanswitch ./cmd/server

# Create runtime image
FROM alpine:latest

# Install ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create app directory
WORKDIR /app

# Create data directory for SQLite database
RUN mkdir -p /app/data && chmod 755 /app/data

# Copy binary from builder
COPY --from=builder /app/deadmanswitch /app/

# Copy web assets
COPY web/templates /app/web/templates
COPY web/static /app/web/static

# Expose the application port
EXPOSE 8080

# Set the entry point
CMD ["/app/deadmanswitch"]