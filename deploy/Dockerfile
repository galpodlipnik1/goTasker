# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies (gcc, musl-dev for CGO)
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# Copy go mod files
COPY app/go.mod app/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY app/ .

# Build the application
ARG TARGETARCH
RUN CGO_ENABLED=1 GOOS=linux GOARCH=$TARGETARCH go build -o gotasker .



# Runtime stage
FROM alpine:latest

# Install curl for healthcheck and sqlite3 runtime dependencies
RUN apk add --no-cache curl ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/gotasker .

# Expose port
EXPOSE 8081

# Run the application
CMD ["./gotasker"]
