# Step 1: Build ARM64 Docker Images - Quick Guide

This is a simplified guide for building ARM64 images for Raspberry Pi deployment.

## Prerequisites

1. **Docker installed** (version 20.10+ with Buildx)
2. **Docker Hub account** (create one at https://hub.docker.com if you don't have one)
3. **Logged into Docker Hub** on your machine

## Quick Method: Use the Automated Script

### Step 1: Make the script executable (if not already)

```bash
chmod +x push-to-dockerhub-arm64.sh
```

### Step 2: Run the script

```bash
./push-to-dockerhub-arm64.sh your-dockerhub-username
```

Replace `your-dockerhub-username` with your actual Docker Hub username.

**Example:**
```bash
./push-to-dockerhub-arm64.sh johndoe
```

The script will:
1. Set up Docker Buildx for ARM64 builds
2. Log you into Docker Hub (if not already logged in)
3. Build all 5 services for ARM64 architecture
4. Push them to Docker Hub

**Time:** This may take 10-30 minutes depending on your internet speed and machine.

## What Gets Built?

1. **mptt-server-ingestor** - MQTT Ingestor Service (Go)
2. **mptt-server-bridge** - MQTT Bridge (Python)
3. **mptt-server-mosquitto** - MQTT Broker
4. **mptt-server-api** - API Service (Go)
5. **mptt-server-frontend** - Frontend (Next.js)

All images will be tagged as `latest` and pushed to:
- `your-username/mptt-server-ingestor:latest`
- `your-username/mptt-server-bridge:latest`
- `your-username/mptt-server-mosquitto:latest`
- `your-username/mptt-server-api:latest`
- `your-username/mptt-server-frontend:latest`

## Verify the Build

After the script completes, verify the images are ARM64:

```bash
docker buildx imagetools inspect your-username/mptt-server-ingestor:latest
```

You should see `linux/arm64` in the output.

## Manual Method (Alternative)

If you prefer to build manually or the script doesn't work:

### 1. Enable Docker Buildx

```bash
docker buildx create --name multiarch --use
docker buildx inspect --bootstrap
```

### 2. Login to Docker Hub

```bash
docker login
```

### 3. Build and Push Each Service

Set your username:
```bash
export DOCKERHUB_USERNAME=your-username
```

**Ingestor:**
```bash
docker buildx build \
    --platform linux/arm64 \
    -t ${DOCKERHUB_USERNAME}/mptt-server-ingestor:latest \
    -f Dockerfile \
    --push \
    .
```

**Bridge:**
```bash
docker buildx build \
    --platform linux/arm64 \
    -t ${DOCKERHUB_USERNAME}/mptt-server-bridge:latest \
    -f ./src/production/MQT.Bridge/Dockerfile \
    --push \
    ./src/production/MQT.Bridge
```

**Mosquitto:**
```bash
docker buildx build \
    --platform linux/arm64 \
    -t ${DOCKERHUB_USERNAME}/mptt-server-mosquitto:latest \
    -f ./src/production/MQT.Mosquitto/Dockerfile \
    --push \
    ./src/production/MQT.Mosquitto
```

**API Service:**
```bash
docker buildx build \
    --platform linux/arm64 \
    -t ${DOCKERHUB_USERNAME}/mptt-server-api:latest \
    -f ./src/production/MQT.ApiService/Dockerfile \
    --push \
    .
```

**Frontend:**
```bash
docker buildx build \
    --platform linux/arm64 \
    --build-arg NEXT_PUBLIC_API_BASE_URL=https://orpheus-networks.com \
    --build-arg NEXT_PUBLIC_READINGS_API_BASE_URL=https://orpheus-networks.com \
    -t ${DOCKERHUB_USERNAME}/mptt-server-frontend:latest \
    -f ./src/production/mqt.frontend/Dockerfile \
    --push \
    ./src/production/mqt.frontend
```

## Troubleshooting

### "Docker Buildx not found"

Update Docker to a newer version, or install buildx:
```bash
# On macOS with Docker Desktop, it's included
# On Linux, you may need to install it separately
```

### "Build takes too long"

ARM64 builds can be slow on AMD64 machines due to emulation. This is normal. The builds will complete, just be patient.

### "Permission denied" or "Cannot connect to Docker"

Make sure Docker is running:
```bash
docker ps
```

If it fails, start Docker Desktop (macOS/Windows) or Docker service (Linux).

### "Authentication failed" when pushing

Make sure you're logged in:
```bash
docker login
```

Enter your Docker Hub username and password.

## Next Steps

After successfully building and pushing the images:

1. ✅ **Step 1 Complete** - ARM64 images are on Docker Hub
2. **Step 2** - Configure DNS for `orpheus-networks.com`
3. **Step 3** - Set up router port forwarding
4. **Step 4** - Deploy on Raspberry Pi

See [DEPLOYMENT.md](DEPLOYMENT.md) for the complete deployment guide.

