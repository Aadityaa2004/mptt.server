# Building ARM64 Docker Images for Raspberry Pi

This guide explains how to build Docker images for ARM64 architecture (Raspberry Pi) and push them to Docker Hub.

## Prerequisites

- Docker installed with Buildx plugin
- Docker Hub account
- Access to the codebase

## Enable Docker Buildx

```bash
# Enable buildx (usually enabled by default in newer Docker versions)
docker buildx version

# Create a new builder instance for multi-arch
docker buildx create --name multiarch --use
docker buildx inspect --bootstrap
```

## Building Multi-Architecture Images

### Option 1: Build for ARM64 Only (Faster)

If you only need ARM64 for Raspberry Pi:

```bash
# Set your Docker Hub username
export DOCKERHUB_USERNAME=your-username

# Build and push Ingestor Service
docker buildx build \
    --platform linux/arm64 \
    -t ${DOCKERHUB_USERNAME}/mptt-server-ingestor:latest \
    -f Dockerfile \
    --push \
    .

# Build and push API Service
docker buildx build \
    --platform linux/arm64 \
    -t ${DOCKERHUB_USERNAME}/mptt-server-api:latest \
    -f ./src/production/MQT.ApiService/Dockerfile \
    --push \
    .

# Build and push Bridge
docker buildx build \
    --platform linux/arm64 \
    -t ${DOCKERHUB_USERNAME}/mptt-server-bridge:latest \
    -f ./src/production/MQT.Bridge/Dockerfile \
    --push \
    ./src/production/MQT.Bridge

# Build and push Mosquitto
docker buildx build \
    --platform linux/arm64 \
    -t ${DOCKERHUB_USERNAME}/mptt-server-mosquitto:latest \
    -f ./src/production/MQT.Mosquitto/Dockerfile \
    --push \
    ./src/production/MQT.Mosquitto

# Build and push Frontend
docker buildx build \
    --platform linux/arm64 \
    --build-arg NEXT_PUBLIC_API_BASE_URL=https://orpheus-networks.com/api \
    --build-arg NEXT_PUBLIC_READINGS_API_BASE_URL=https://orpheus-networks.com/api \
    -t ${DOCKERHUB_USERNAME}/mptt-server-frontend:latest \
    -f ./src/production/mqt.frontend/Dockerfile \
    --push \
    ./src/production/mqt.frontend
```

### Option 2: Build Multi-Arch (AMD64 + ARM64)

If you want to support both architectures:

```bash
# Set your Docker Hub username
export DOCKERHUB_USERNAME=your-username

# Build and push Ingestor Service (multi-arch)
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -t ${DOCKERHUB_USERNAME}/mptt-server-ingestor:latest \
    -f Dockerfile \
    --push \
    .

# Build and push API Service (multi-arch)
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -t ${DOCKERHUB_USERNAME}/mptt-server-api:latest \
    -f ./src/production/MQT.ApiService/Dockerfile \
    --push \
    .

# Build and push Bridge (multi-arch)
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -t ${DOCKERHUB_USERNAME}/mptt-server-bridge:latest \
    -f ./src/production/MQT.Bridge/Dockerfile \
    --push \
    ./src/production/MQT.Bridge

# Build and push Mosquitto (multi-arch)
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -t ${DOCKERHUB_USERNAME}/mptt-server-mosquitto:latest \
    -f ./src/production/MQT.Mosquitto/Dockerfile \
    --push \
    ./src/production/MQT.Mosquitto

# Build and push Frontend (multi-arch)
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --build-arg NEXT_PUBLIC_API_BASE_URL=https://orpheus-networks.com/api \
    --build-arg NEXT_PUBLIC_READINGS_API_BASE_URL=https://orpheus-networks.com/api \
    -t ${DOCKERHUB_USERNAME}/mptt-server-frontend:latest \
    -f ./src/production/mqt.frontend/Dockerfile \
    --push \
    ./src/production/mqt.frontend
```

## Using the Updated push-to-dockerhub.sh Script

You can update `push-to-dockerhub.sh` to support ARM64 builds. Here's an example modification:

```bash
# Add this at the beginning of the script
docker buildx create --name multiarch --use 2>/dev/null || docker buildx use multiarch
docker buildx inspect --bootstrap

# Then modify build commands to use buildx:
docker buildx build \
    --platform linux/arm64 \
    -t "${INGESTOR_IMAGE}" \
    -f Dockerfile \
    --push \
    .
```

## Verifying ARM64 Images

After pushing, verify the architecture:

```bash
# Check image architecture
docker buildx imagetools inspect ${DOCKERHUB_USERNAME}/mptt-server-ingestor:latest
```

## Notes

- **Distroless Images**: The Go services use `gcr.io/distroless/base-debian12` which supports ARM64
- **Node.js Images**: `node:20-alpine` supports ARM64 natively
- **Python Images**: `python:3.11-slim` supports ARM64 natively
- **Mosquitto**: `eclipse-mosquitto:2` supports ARM64 natively

All base images used in the Dockerfiles support ARM64, so the builds should work correctly.

## Troubleshooting

### Build Fails with "exec format error"

This usually means you're trying to run an AMD64 image on ARM64. Make sure you:
1. Built the image with `--platform linux/arm64`
2. Pushed the ARM64 version to Docker Hub
3. Pulled the ARM64 version on Raspberry Pi

### Build Takes Too Long

ARM64 builds can be slower, especially if using emulation. Consider:
- Using a Raspberry Pi or ARM64 build server for native builds
- Using GitHub Actions or GitLab CI with ARM64 runners
- Building locally on an ARM64 machine

### Missing Dependencies

Some dependencies might need ARM64-specific builds. Check:
- Go modules compile correctly for ARM64
- Node.js packages have ARM64 binaries
- Python packages have ARM64 wheels

