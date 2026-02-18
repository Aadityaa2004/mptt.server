# Multi-stage build for Go Email Alert Service
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

# Use TARGETARCH from buildx when set; otherwise use build host arch (so native build on Pi = arm64)
ARG TARGETARCH
RUN export GOARCH=${TARGETARCH:-$(go env GOARCH)} && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /bin/email-service \
    ./src/production/MQT.EmailService

# Stage 2: Runtime (distroless)
FROM gcr.io/distroless/base-debian12

# Set timezone
ENV TZ=Etc/UTC

# Set working directory
WORKDIR /app

# Copy the binary from build stage
COPY --from=build /bin/email-service /email-service

# Copy timezone data
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose port
EXPOSE 9004

# Set entrypoint
ENTRYPOINT ["/email-service"]
