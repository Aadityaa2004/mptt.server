# Multi-stage build for Go MQTT Ingestor Service
# Stage 1: Build
FROM golang:1.25-alpine AS build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /src

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /bin/ingestor \
    ./src/production/MQT.Startup

# Stage 2: Runtime (distroless)
FROM gcr.io/distroless/base-debian12

# Set timezone
ENV TZ=Etc/UTC

# Set working directory
WORKDIR /app

# Copy the binary from build stage
COPY --from=build /bin/ingestor /ingestor

# Copy timezone data
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose port
EXPOSE 9002

# Set entrypoint
ENTRYPOINT ["/ingestor"]
