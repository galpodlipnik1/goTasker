# Build stage
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY app/go.mod app/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY app/ .

# Build the application
ARG TARGETARCH
ARG TARGETOS
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o gotasker .



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
