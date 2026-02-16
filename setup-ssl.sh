#!/bin/bash

# SSL Certificate Setup Script for Let's Encrypt
# This script sets up SSL certificates for orpheus-networks.com

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}SSL Certificate Setup${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Check if .env.production exists
if [ ! -f ".env.production" ]; then
    echo -e "${RED}Error: .env.production file not found!${NC}"
    exit 1
fi

# Load environment variables
set -a
source .env.production
set +a

DOMAIN=${DOMAIN:-orpheus-networks.com}
EMAIL=${SSL_EMAIL:-admin@orpheus-networks.com}

echo -e "${YELLOW}Domain: ${DOMAIN}${NC}"
echo -e "${YELLOW}Email: ${EMAIL}${NC}\n"

# Check if Docker is running
if ! docker ps &> /dev/null; then
    echo -e "${RED}Error: Docker is not running${NC}"
    exit 1
fi

# Check if Nginx container exists and is running
if ! docker ps | grep -q nginx-proxy; then
    echo -e "${YELLOW}Starting Nginx container first...${NC}"
    
    # Use docker compose (newer) or docker-compose (older)
    if docker compose version &> /dev/null; then
        DOCKER_COMPOSE="docker compose"
    else
        DOCKER_COMPOSE="docker-compose"
    fi
    
    # Create network if it doesn't exist
    docker network create mqtt-network 2>/dev/null || true
    
    # Start Nginx
    $DOCKER_COMPOSE -f docker-compose.nginx.yml up -d
    
    echo -e "${GREEN}✓ Nginx started${NC}\n"
    sleep 5
fi

# Check if certificates already exist
if docker exec nginx-proxy test -f /etc/letsencrypt/live/${DOMAIN}/fullchain.pem 2>/dev/null; then
    echo -e "${YELLOW}SSL certificates already exist for ${DOMAIN}${NC}"
    read -p "Do you want to renew them? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${GREEN}Keeping existing certificates${NC}"
        exit 0
    fi
fi

echo -e "${BLUE}Requesting SSL certificate from Let's Encrypt...${NC}"
echo -e "${YELLOW}Make sure:${NC}"
echo -e "  1. Domain ${DOMAIN} DNS points to this server's public IP"
echo -e "  2. Ports 80 and 443 are open and forwarded to this Raspberry Pi"
echo -e "  3. You have access to ${EMAIL} for certificate notifications\n"

read -p "Press Enter to continue or Ctrl+C to cancel..."

# Request certificate using certbot in Nginx container
echo -e "\n${BLUE}Running certbot...${NC}"

docker exec nginx-proxy certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email ${EMAIL} \
    --agree-tos \
    --no-eff-email \
    --force-renewal \
    -d ${DOMAIN} \
    -d www.${DOMAIN} || {
    echo -e "\n${RED}Certificate generation failed!${NC}"
    echo -e "${YELLOW}Common issues:${NC}"
    echo -e "  - DNS not pointing to this server"
    echo -e "  - Port 80 not accessible from internet"
    echo -e "  - Firewall blocking port 80"
    exit 1
}

echo -e "\n${GREEN}✓ SSL certificates generated successfully!${NC}\n"

# Reload Nginx to use new certificates
echo -e "${BLUE}Reloading Nginx...${NC}"
docker exec nginx-proxy nginx -s reload
echo -e "${GREEN}✓ Nginx reloaded${NC}\n"

# Set up automatic renewal
echo -e "${BLUE}Setting up automatic certificate renewal...${NC}"

# Create renewal script
cat > renew-certs.sh << 'EOF'
#!/bin/bash
docker exec nginx-proxy certbot renew --quiet
docker exec nginx-proxy nginx -s reload
EOF

chmod +x renew-certs.sh

# Add to crontab (runs twice daily)
(crontab -l 2>/dev/null | grep -v "renew-certs.sh"; echo "0 0,12 * * * $(pwd)/renew-certs.sh >> /var/log/certbot-renewal.log 2>&1") | crontab -

echo -e "${GREEN}✓ Automatic renewal configured (runs twice daily)${NC}\n"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}SSL Setup Completed!${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${YELLOW}Test your SSL certificate:${NC}"
echo -e "  https://www.ssllabs.com/ssltest/analyze.html?d=${DOMAIN}"
echo -e "\n${YELLOW}Access your site:${NC}"
echo -e "  https://${DOMAIN}"

