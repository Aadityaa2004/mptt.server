#!/bin/bash

# Script to build and push Docker images to Docker Hub
# Usage: ./push-to-dockerhub.sh [your-dockerhub-username]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Image tags (you can customize these)
VERSION=${VERSION:-latest}
PROJECT_NAME="mptt-server"

echo -e "${GREEN}Building and pushing images to Docker Hub as ${DOCKERHUB_USERNAME}${NC}\n"

# Login to Docker Hub
echo -e "${YELLOW}Logging in to Docker Hub...${NC}"
docker login

# Build and push MQTT Ingestor Service (main service)
echo -e "\n${GREEN}[1/6] Building MQTT Ingestor Service...${NC}"
INGESTOR_IMAGE="${DOCKERHUB_USERNAME}/${PROJECT_NAME}-ingestor:${VERSION}"
docker build -t "${INGESTOR_IMAGE}" -f Dockerfile .
docker tag "${INGESTOR_IMAGE}" "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-ingestor:latest"
echo -e "${GREEN}Pushing MQTT Ingestor Service...${NC}"
docker push "${INGESTOR_IMAGE}"
docker push "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-ingestor:latest"

# Build and push MQTT Bridge
echo -e "\n${GREEN}[2/6] Building MQTT Bridge...${NC}"
BRIDGE_IMAGE="${DOCKERHUB_USERNAME}/${PROJECT_NAME}-bridge:${VERSION}"
docker build -t "${BRIDGE_IMAGE}" -f ./src/production/MQT.Bridge/Dockerfile ./src/production/MQT.Bridge
docker tag "${BRIDGE_IMAGE}" "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-bridge:latest"
echo -e "${GREEN}Pushing MQTT Bridge...${NC}"
docker push "${BRIDGE_IMAGE}"
docker push "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-bridge:latest"

# Build and push MQTT Mosquitto
echo -e "\n${GREEN}[3/6] Building MQTT Mosquitto...${NC}"
MOSQUITTO_IMAGE="${DOCKERHUB_USERNAME}/${PROJECT_NAME}-mosquitto:${VERSION}"
docker build -t "${MOSQUITTO_IMAGE}" -f ./src/production/MQT.Mosquitto/Dockerfile ./src/production/MQT.Mosquitto
docker tag "${MOSQUITTO_IMAGE}" "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-mosquitto:latest"
echo -e "${GREEN}Pushing MQTT Mosquitto...${NC}"
docker push "${MOSQUITTO_IMAGE}"
docker push "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-mosquitto:latest"

# Build and push API Service
echo -e "\n${GREEN}[4/6] Building API Service...${NC}"
API_IMAGE="${DOCKERHUB_USERNAME}/${PROJECT_NAME}-api:${VERSION}"
docker build -t "${API_IMAGE}" -f ./src/production/MQT.ApiService/Dockerfile .
docker tag "${API_IMAGE}" "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-api:latest"
echo -e "${GREEN}Pushing API Service...${NC}"
docker push "${API_IMAGE}"
docker push "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-api:latest"

# Build and push Email Service
echo -e "\n${GREEN}[5/6] Building MQT Email Service...${NC}"
EMAIL_IMAGE="${DOCKERHUB_USERNAME}/${PROJECT_NAME}-email:${VERSION}"
docker build -t "${EMAIL_IMAGE}" -f ./src/production/MQT.EmailService/Dockerfile .
docker tag "${EMAIL_IMAGE}" "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-email:latest"
echo -e "${GREEN}Pushing MQT Email Service...${NC}"
docker push "${EMAIL_IMAGE}"
docker push "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-email:latest"

# Build and push Frontend (production URLs baked in for RPi/production use)
echo -e "\n${GREEN}[6/6] Building Frontend...${NC}"
FRONTEND_IMAGE="${DOCKERHUB_USERNAME}/${PROJECT_NAME}-frontend:${VERSION}"
docker build -t "${FRONTEND_IMAGE}" \
  --build-arg NEXT_PUBLIC_API_BASE_URL="${NEXT_PUBLIC_API_BASE_URL:-https://orpheus-networks.com}" \
  --build-arg NEXT_PUBLIC_READINGS_API_BASE_URL="${NEXT_PUBLIC_READINGS_API_BASE_URL:-https://orpheus-networks.com}" \
  -f ./src/production/mqt.frontend/Dockerfile ./src/production/mqt.frontend
docker tag "${FRONTEND_IMAGE}" "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-frontend:latest"
echo -e "${GREEN}Pushing Frontend...${NC}"
docker push "${FRONTEND_IMAGE}"
docker push "${DOCKERHUB_USERNAME}/${PROJECT_NAME}-frontend:latest"

echo -e "\n${GREEN}✓ All images pushed successfully!${NC}\n"
echo -e "${YELLOW}Images pushed:${NC}"
echo "  - ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-ingestor:${VERSION}"
echo "  - ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-bridge:${VERSION}"
echo "  - ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-mosquitto:${VERSION}"
echo "  - ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-api:${VERSION}"
echo "  - ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-email:${VERSION}"
echo "  - ${DOCKERHUB_USERNAME}/${PROJECT_NAME}-frontend:${VERSION}"
