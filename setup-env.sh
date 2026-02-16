#!/bin/bash

# Interactive environment setup script
# Helps user create .env.production from template

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Environment Setup Wizard${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Check if template exists (try both with and without dot prefix)
ENV_TEMPLATE=""
if [ -f ".env.production.example" ]; then
    ENV_TEMPLATE=".env.production.example"
elif [ -f "env.production.example" ]; then
    ENV_TEMPLATE="env.production.example"
else
    echo -e "${RED}Error: .env.production.example not found!${NC}"
    exit 1
fi

# Check if .env.production already exists
if [ -f ".env.production" ]; then
    echo -e "${YELLOW}Warning: .env.production already exists!${NC}"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Cancelled. Keeping existing .env.production${NC}"
        exit 0
    fi
fi

# Copy template
cp "$ENV_TEMPLATE" .env.production
echo -e "${GREEN}✓ Created .env.production from template${NC}\n"

# Function to generate random string
generate_secret() {
    openssl rand -base64 32 | tr -d "=+/" | cut -c1-32
}

# Function to prompt for value with default
prompt_value() {
    local var_name=$1
    local prompt_text=$2
    local default_value=$3
    local is_secret=${4:-false}
    
    if [ "$is_secret" = true ]; then
        read -sp "${prompt_text} [default: ${default_value}]: " value
        echo
    else
        read -p "${prompt_text} [default: ${default_value}]: " value
    fi
    
    if [ -z "$value" ]; then
        value=$default_value
    fi
    
    # Update the file
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s|^${var_name}=.*|${var_name}=${value}|" .env.production
    else
        # Linux
        sed -i "s|^${var_name}=.*|${var_name}=${value}|" .env.production
    fi
}

echo -e "${BLUE}Please provide the following information:${NC}\n"

# Docker Hub Username
prompt_value "DOCKERHUB_USERNAME" "Docker Hub Username" "CHANGE_ME"

# Database Configuration
echo -e "\n${YELLOW}Database Configuration:${NC}"
prompt_value "POSTGRES_USER" "PostgreSQL Username" "iot_user"
prompt_value "POSTGRES_PASSWORD" "PostgreSQL Password (will be hidden)" "CHANGE_ME_STRONG_PASSWORD" true
prompt_value "POSTGRES_DB" "PostgreSQL Database Name" "iot"

# Generate JWT Secret
echo -e "\n${YELLOW}Generating JWT Secret...${NC}"
JWT_SECRET=$(generate_secret)
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s|^JWT_SECRET_KEY=.*|JWT_SECRET_KEY=${JWT_SECRET}|" .env.production
else
    sed -i "s|^JWT_SECRET_KEY=.*|JWT_SECRET_KEY=${JWT_SECRET}|" .env.production
fi
echo -e "${GREEN}✓ Generated JWT_SECRET_KEY${NC}"

# Generate Internal API Secret
echo -e "${YELLOW}Generating Internal API Secret...${NC}"
INTERNAL_SECRET=$(generate_secret)
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s|^INTERNAL_API_SECRET=.*|INTERNAL_API_SECRET=${INTERNAL_SECRET}|" .env.production
else
    sed -i "s|^INTERNAL_API_SECRET=.*|INTERNAL_API_SECRET=${INTERNAL_SECRET}|" .env.production
fi
echo -e "${GREEN}✓ Generated INTERNAL_API_SECRET${NC}"

# Admin Account
echo -e "\n${YELLOW}Admin Account Configuration:${NC}"
prompt_value "ADMIN_USERNAME" "Admin Username" "admin"
prompt_value "ADMIN_EMAIL" "Admin Email" "admin@orpheus-networks.com"
prompt_value "ADMIN_PASSWORD" "Admin Password (will be hidden)" "CHANGE_ME_STRONG_ADMIN_PASSWORD" true

# MQTT Broker
echo -e "\n${YELLOW}MQTT Broker Configuration:${NC}"
prompt_value "BROKER_USER" "MQTT Broker Username" "CHANGE_ME_MQTT_USERNAME"
prompt_value "BROKER_PASS" "MQTT Broker Password (will be hidden)" "CHANGE_ME_MQTT_PASSWORD" true

# External APIs
echo -e "\n${YELLOW}External API Keys (optional):${NC}"
prompt_value "OPENWEATHER_API_KEY" "OpenWeather API Key (optional)" ""

# Set file permissions
chmod 600 .env.production
echo -e "\n${GREEN}✓ Set file permissions to 600${NC}"

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}Environment Setup Completed!${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${YELLOW}Next steps:${NC}"
echo -e "1. Review .env.production and update any remaining values"
echo -e "2. Run: ./deploy-rpi.sh to deploy services"
echo -e "3. Run: ./setup-ssl.sh to set up SSL certificates"

