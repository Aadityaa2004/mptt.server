#!/bin/bash

# Deployment script for Raspberry Pi production environment
# This script validates environment, pulls images, and starts services

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

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Raspberry Pi Production Deployment${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Check if .env.production exists
if [ ! -f ".env.production" ]; then
    echo -e "${RED}Error: .env.production file not found!${NC}"
    ENV_TEMPLATE=""
    if [ -f ".env.production.example" ]; then
        ENV_TEMPLATE=".env.production.example"
    elif [ -f "env.production.example" ]; then
        ENV_TEMPLATE="env.production.example"
    fi
    if [ -n "$ENV_TEMPLATE" ]; then
        echo -e "${YELLOW}Please copy $ENV_TEMPLATE to .env.production and fill in your values.${NC}"
    else
        echo -e "${YELLOW}Please create .env.production with your configuration values.${NC}"
    fi
    echo -e "${YELLOW}Copy env.production.example to .env.production and fill in your values.${NC}"
    exit 1
fi

# Check file permissions
ENV_PERMS=$(stat -c "%a" .env.production 2>/dev/null || stat -f "%OLp" .env.production 2>/dev/null)
if [ "$ENV_PERMS" != "600" ] && [ "$ENV_PERMS" != "0600" ]; then
    echo -e "${YELLOW}Warning: .env.production permissions are $ENV_PERMS (should be 600)${NC}"
    echo -e "${YELLOW}Setting permissions to 600...${NC}"
    chmod 600 .env.production
    echo -e "${GREEN}✓ Permissions updated${NC}\n"
fi

# Load environment variables
set -a
source .env.production
set +a

# Validate required variables
REQUIRED_VARS=(
    "DOCKERHUB_USERNAME"
    "POSTGRES_USER"
    "POSTGRES_PASSWORD"
    "JWT_SECRET_KEY"
    "INTERNAL_API_SECRET"
    "ADMIN_PASSWORD"
)

MISSING_VARS=()
for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ] || [[ "${!var}" == *"CHANGE_ME"* ]]; then
        MISSING_VARS+=("$var")
    fi
done

if [ ${#MISSING_VARS[@]} -ne 0 ]; then
    echo -e "${RED}Error: Missing or unset required environment variables:${NC}"
    for var in "${MISSING_VARS[@]}"; do
        echo -e "  - ${RED}$var${NC}"
    done
    echo -e "\n${YELLOW}Please update .env.production with proper values.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Environment file validated${NC}\n"

# Check Docker and Docker Compose
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}Error: Docker Compose is not installed${NC}"
    exit 1
fi

# Use docker compose (newer) or docker-compose (older)
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

echo -e "${BLUE}Pulling Docker images from Docker Hub...${NC}\n"

# Pull images
$DOCKER_COMPOSE -f docker-compose.rpi.yml pull

echo -e "\n${GREEN}✓ Images pulled successfully${NC}\n"

# Create network if it doesn't exist
echo -e "${BLUE}Setting up Docker network...${NC}"
docker network create mqtt-network 2>/dev/null || true
echo -e "${GREEN}✓ Network ready${NC}\n"

# Start services
echo -e "${BLUE}Starting services...${NC}\n"
$DOCKER_COMPOSE -f docker-compose.rpi.yml up -d

echo -e "\n${GREEN}✓ Services started${NC}\n"

# Wait for services to be healthy
echo -e "${BLUE}Waiting for services to be healthy...${NC}"
sleep 10

# Check service health
echo -e "\n${BLUE}Checking service health...${NC}"
docker ps --format "table {{.Names}}\t{{.Status}}"

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}Deployment completed successfully!${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${YELLOW}Next steps:${NC}"
echo -e "1. Verify Cloudflare Tunnel is running: docker compose -f docker-compose.rpi.yml logs cloudflared"
echo -e "2. Configure tunnel hostname in Cloudflare dashboard: orpheus-networks.com -> http://nginx:80"
echo -e "3. Test external access: https://orpheus-networks.com"
echo -e "\n${BLUE}To view logs:${NC}"
echo -e "  $DOCKER_COMPOSE -f docker-compose.rpi.yml logs -f"
echo -e "\n${BLUE}To stop services:${NC}"
echo -e "  $DOCKER_COMPOSE -f docker-compose.rpi.yml down"

