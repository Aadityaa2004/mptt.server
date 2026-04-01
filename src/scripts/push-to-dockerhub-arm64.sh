#!/bin/bash

# Script to build and push ARM64 Docker images to Docker Hub for Raspberry Pi
# Usage: ./push-to-dockerhub-arm64.sh [your-dockerhub-username]

set -e

# Resolve script directory and project root (works when run from src/scripts or project root)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
cd "${ROOT_DIR}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get Docker Hub username from argument or prompt
DOCKERHUB_USERNAME=${1:-}
if [ -z "$DOCKERHUB_USERNAME" ]; then
    echo -e "${YELLOW}Enter your Docker Hub username:${NC}"
    read DOCKERHUB_USERNAME
fi

if [ -z "$DOCKERHUB_USERNAME" ]; then
    echo -e "${RED}Error: Docker Hub username is required${NC}"
    exit 1
fi

# Image tags
VERSION=${VERSION:-latest}
PROJECT_NAME="mptt-server"
PLATFORM="linux/arm64"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Building ARM64 Images for Raspberry Pi${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Check if buildx is available
if ! docker buildx version &> /dev/null; then
    echo -e "${RED}Error: Docker Buildx is not available${NC}"
    echo -e "${YELLOW}Please install Docker Buildx or update Docker to a newer version${NC}"
    exit 1
fi

# Setup buildx builder
echo -e "${YELLOW}Setting up Docker Buildx...${NC}"
docker buildx create --name multiarch --use 2>/dev/null || docker buildx use multiarch
docker buildx inspect --bootstrap
echo -e "${GREEN}✓ Buildx ready${NC}\n"

# Login to Docker Hub
echo -e "${YELLOW}Logging in to Docker Hub...${NC}"
docker login
echo -e "${GREEN}✓ Logged in${NC}\n"

# Build and push MQTT Ingestor Service
echo -e "\n${GREEN}[1/6] Building MQTT Ingestor Service (ARM64)...${NC}"
INGESTOR_IMAGE="${DOCKERHUB_USERNAME}/${PROJECT_NAME}-ingestor:${VERSION}"
docker buildx build \
    --platform ${PLATFORM} \
    -t "${INGESTOR_IMAGE}" \
    -t "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-ingestor:latest" \
    -f Dockerfile \
    --push \
    .
echo -e "${GREEN}✓ Ingestor Service pushed${NC}"

# Build and push MQTT Bridge
echo -e "\n${GREEN}[2/6] Building MQTT Bridge (ARM64)...${NC}"
BRIDGE_IMAGE="${DOCKERHUB_USERNAME}/${PROJECT_NAME}-bridge:${VERSION}"
docker buildx build \
    --platform ${PLATFORM} \
    -t "${BRIDGE_IMAGE}" \
    -t "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-bridge:latest" \
    -f ./src/production/MQT.Bridge/Dockerfile \
    --push \
    ./src/production/MQT.Bridge
echo -e "${GREEN}✓ MQTT Bridge pushed${NC}"

# Build and push MQTT Mosquitto
echo -e "\n${GREEN}[3/6] Building MQTT Mosquitto (ARM64)...${NC}"
MOSQUITTO_IMAGE="${DOCKERHUB_USERNAME}/${PROJECT_NAME}-mosquitto:${VERSION}"
docker buildx build \
    --platform ${PLATFORM} \
    -t "${MOSQUITTO_IMAGE}" \
    -t "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-mosquitto:latest" \
    -f ./src/production/MQT.Mosquitto/Dockerfile \
    --push \
    ./src/production/MQT.Mosquitto
echo -e "${GREEN}✓ MQTT Mosquitto pushed${NC}"

# Build and push API Service
echo -e "\n${GREEN}[4/6] Building API Service (ARM64)...${NC}"
API_IMAGE="${DOCKERHUB_USERNAME}/${PROJECT_NAME}-api:${VERSION}"
docker buildx build \
    --platform ${PLATFORM} \
    -t "${API_IMAGE}" \
    -t "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-api:latest" \
    -f ./src/production/MQT.ApiService/Dockerfile \
    --push \
    .
echo -e "${GREEN}✓ API Service pushed${NC}"

# Build and push Email Service (same stack as docker-compose.rpi.yml)
echo -e "\n${GREEN}[5/6] Building MQT Email Service (ARM64)...${NC}"
EMAIL_IMAGE="${DOCKERHUB_USERNAME}/${PROJECT_NAME}-email:${VERSION}"
docker buildx build \
    --platform ${PLATFORM} \
    -t "${EMAIL_IMAGE}" \
    -t "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-email:latest" \
    -f ./src/production/MQT.EmailService/Dockerfile \
    --push \
    .
echo -e "${GREEN}✓ MQT Email Service pushed${NC}"

# Build and push Frontend
echo -e "\n${GREEN}[6/6] Building Frontend (ARM64)...${NC}"
FRONTEND_IMAGE="${DOCKERHUB_USERNAME}/${PROJECT_NAME}-frontend:${VERSION}"
docker buildx build \
    --platform ${PLATFORM} \
    --build-arg NEXT_PUBLIC_API_BASE_URL=https://orpheus-networks.com/api \
    --build-arg NEXT_PUBLIC_READINGS_API_BASE_URL=https://orpheus-networks.com/api \
    -t "${FRONTEND_IMAGE}" \
    -t "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-frontend:latest" \
    -f ./src/production/mqt.frontend/Dockerfile \
    --push \
    ./src/production/mqt.frontend
echo -e "${GREEN}✓ Frontend pushed${NC}"

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}✓ All ARM64 images pushed successfully!${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${YELLOW}Images pushed:${NC}"
echo "  - ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-ingestor:${VERSION}"
echo "  - ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-bridge:${VERSION}"
echo "  - ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-mosquitto:${VERSION}"
echo "  - ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-api:${VERSION}"
echo "  - ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-email:${VERSION}"
echo "  - ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-frontend:${VERSION}"

echo -e "\n${BLUE}Verify architecture (example):${NC}"
echo "  docker buildx imagetools inspect ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-ingestor:latest"
echo "  docker buildx imagetools inspect ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-email:latest"

